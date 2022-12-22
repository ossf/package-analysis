# Queries

This document contains useful queries to run in Google Cloud BigQuery.

To use these queries:

1. Visit the BigQuery [SQL Workspace](https://console.cloud.google.com/bigquery)
in the GCP console.
1. Select or create a new editor.
1. Copy and paste the query in and adjust the variables as needed.
1. Click "Run" to execute the query.
1. Click "View Results" if needed when the query is complete.

**NOTE:** executing these queries may incur charges against your GCP project.

## DNS Queries

DNS queries requested over `LOOKBACK_DAYS`, ordered by domain.

Change `install` to `import` to return results for different phases.

```sql
DECLARE LOOKBACK_DAYS INT64 DEFAULT 2;
DECLARE PHASE DEFAULT "import";

WITH dns_queries AS (
    SELECT Package.ecosystem, Package.name, Package.version, queries.Hostname as Hostname, QueryType FROM
        `ossf-malware-analysis.packages`.analysis_for_phase(PHASE) AS t,
        t.Result.DNS as DNS,
        UNNEST(DNS.Queries) AS queries,
        UNNEST(queries.Types) AS QueryType
    WHERE 
        DNS.Class = 'IN'
        AND TIMESTAMP_ADD(CreatedTimestamp, INTERVAL LOOKBACK_DAYS DAY) > CURRENT_TIMESTAMP()
)
SELECT
  ecosystem,
  name,
  Hostname,
  QueryType,
  ARRAY_AGG(version),
  ARRAY_TO_STRING(ARRAY_REVERSE(SPLIT(Hostname, ".")), ".") AS rev
FROM
  dns_queries
WHERE
  /* Ignore hostname lookups for "A" (IPv4 addresses) and "AAAA" (IPv6 addresses)
     records with only one part to limit noise */
  (QueryType NOT IN ('A', 'AAAA') OR ARRAY_LENGTH(SPLIT(Hostname, ".")) >= 2) AND
  /* Ignore "safe" jostnames that are used frequently */
  Hostname NOT IN ('gcr.io.default.svc.cluster.local', 'storage.googleapis.com.cluster.local',
                   'storage.googleapis.com.svc.cluster.local', 'gcr.io.cluster.local',
                   'gcr.io.google.internal', 'storage.googleapis.com.default.svc.cluster.local',
                   'storage.googleapis.com.us-central1-c.c.ossf-malware-analysis.internal',
                   'gcr.io.default.svc.cluster.local', 'storage.googleapis.com.google.internal',
                   'storage.googleapis.com.c.ossf-malware-analysis.internal',
                   'gcr.io.svc.cluster.internal', 'gcr.io.us-central1-c.c.ossf-malware-analysis.internal',
                   'gcr.io.svc.cluster.local', 'gcr.io.c.ossf-malware-analysis.internal',
                   'registry.npmjs.org', 'storage.googleapis.com', 'files.pythonhosted.org',
                   'index.rubygems.org', 'pypi.org', 'rubygems.org', 'packagist.org',
                   'repo.packagist.org', 'api.github.com', 'github.com', 'bitbucket.org',
                   'codeload.github.com', 'objects.githubusercontent.com', 'opencollective.com',
                   'crates.io', 'static.crates.io')
GROUP BY ecosystem, name, Hostname, QueryType
ORDER by rev ASC, name ASC, QueryType ASC;
```

## Opened Sockets

Returns the last 1000 sockets opened during the specified phase.

Hides hostnames used commonly.

```sql
DECLARE COMMON_HOSTNAME_THRESHOLD DEFAULT 10000;
DECLARE PHASE DEFAULT "import";

CREATE TEMP FUNCTION dedupe(sockets ARRAY<STRUCT<Hostnames ARRAY<STRING>, Port INTEGER, Address STRING>>)
RETURNS ARRAY<STRUCT<Hostnames ARRAY<STRING>, Port INTEGER, Address STRING>>
LANGUAGE js AS r"""
  let map = {};
  for (let sk of sockets) {
      map[sk.Address] = sk;
  }

  return Object.values(map).sort(function(a, b) {
      if (a.Address > b.Address) {
          return 1;
      } else if (a.Address < b.Address) {
          return -1;
      } else {
          return 0;
      }
  } );
""";

SELECT
  CreatedTimestamp as t,
  Package.ecosystem,
  Package.name,
  Package.version,
  dedupe(ARRAY_AGG(sockets)) as sockets
FROM
  `ossf-malware-analysis.packages`.analysis_for_phase(PHASE) AS t,
  t.Result.Sockets as sockets
WHERE
  (
    (
      ARRAY_LENGTH(sockets.Hostnames) = 0 AND
      /* mitigate bug with missing hostnames by looking for other matching entries. */
      NOT EXISTS (
        SELECT sk
        FROM
          `ossf-malware-analysis.packages`.analysis_for_phase(PHASE) AS t,
          t.Result.Sockets AS sk
        WHERE ARRAY_LENGTH(sk.Hostnames) != 0 AND sk.Address = sockets.Address
      )
    )
    /* uncommon hostnames */
    OR (
      SELECT COUNT(hn)
      FROM
        `ossf-malware-analysis.packages`.analysis_for_phase(PHASE) AS t,
        UNNEST(t.Result.Sockets) AS sk,
        UNNEST(sk.Hostnames) AS hn
      WHERE hn IN UNNEST(sockets.Hostnames)
    ) < COMMON_HOSTNAME_THRESHOLD
  )
  AND TIMESTAMP_ADD(CreatedTimestamp, INTERVAL 30 DAY) > CURRENT_TIMESTAMP()
  AND sockets.Address != "::1"
  AND sockets.Address != "127.0.0.1"
  AND sockets.Address != "8.8.8.8"
  AND sockets.Address != "8.8.4.4"
  AND sockets.Address != ""
  AND NOT sockets.Address LIKE "10.%" /* Private Class A (10.x.x.x) */
  AND (  /* Private Class B (172.16.x.x -> 172.31.x.x) */
    CONTAINS_SUBSTR(sockets.Address, ':')
    OR NET.IPV4_TO_INT64(NET.IP_FROM_STRING(sockets.Address)) < 0xAC100000
    OR NET.IPV4_TO_INT64(NET.IP_FROM_STRING(sockets.Address)) > 0xAC1FFFFF)
  AND NOT sockets.Address LIKE "192.168.%" /* Private Class C (192.168.x.x) */
GROUP BY t, Package.ecosystem, Package.name, Package.version
ORDER by t DESC, Package.name
LIMIT 1000;
```

## Exec'd Commands

Shows commands executed, sorted by command the number of occurances.

A large part of the SQL attempts to filter out noise, which it does with limited success.

Currently set to show 2 days for PyPI during the `import` phase.

TODO: clean up the JavaScript.

```sql
DECLARE PKG_ECOSYSTEM STRING DEFAULT 'pypi';
DECLARE LOOKBACK_DAYS INT64 DEFAULT 2;
DECLARE PHASE DEFAULT "import";

CREATE TEMP FUNCTION is_known_good(pkg STRING, ver STRING, c STRUCT<Environment ARRAY<STRING>, Command ARRAY<STRING>>, e STRING)
RETURNS BOOL 
LANGUAGE js
AS r"""
    let command = c.Command;
    let singleCmd = command.join(" ");
    if (singleCmd == "sleep 30m" || singleCmd == "npm init --force" || singleCmd == "node /usr/local/bin/npm init --force" || singleCmd == "sh -c node -e \"try{require('./postinstall')}catch(e){}\""
       || singleCmd == "node -e try{require('./postinstall')}catch(e){}" || singleCmd == "/bin/sh -c node-gyp-build-test" || singleCmd == "node /app/node_modules/.bin/node-gyp-build-test"
       || singleCmd == "node /app/node_modules/.bin/node-gyp-build" || singleCmd == "node /usr/local/bin/npm run rebuild" || singleCmd == "node-gyp rebuild" || singleCmd == "node-gyp-build-test"
       || singleCmd == "node-gyp-build" || singleCmd == "node /usr/local/lib/node_modules/npm/node_modules/node-gyp/bin/node-gyp.js rebuild" || singleCmd == "npm run rebuild"
       || singleCmd == "sh -c node-gyp rebuild || exit 0" || singleCmd == "sh -c node-gyp rebuild" || singleCmd == "sh -c node-gyp-build || exit 0" || singleCmd == "sh -c node-gyp-build"
       || singleCmd == "sh /usr/local/lib/node_modules/npm/node_modules/@npmcli/run-script/lib/node-gyp-bin/node-gyp rebuild"
       || singleCmd == "/app/node_modules/esbuild/bin/esbuild --version" || singleCmd == "node /usr/local/bin/npm run build" || singleCmd == "npm run build"
       || singleCmd == "python -c import sys; print(sys.executable);" || singleCmd == "python2 -c import sys; print(sys.executable);"
       || singleCmd == "python3 -c import sys; print(sys.executable);" || singleCmd == "node-gyp rebuild --release"
       || singleCmd == "node install.js" || singleCmd == "sh -c node install.js"
       || singleCmd.startsWith("/bin/sh -c \"/usr/local/bin/node\" /tmp/shelljs_")
       || singleCmd.startsWith("/bin/sh -c \"/usr/local/bin/node\" \"/tmp/shelljs_")) {
        return true;
    }
    if (command[0] == "node") {
        if (command[1] == "/usr/local/bin/analyze.js") {
            return true;
        }
        if (command[1] == "/usr/local/bin/npm" && command[2] == "install") {
            if (command[3] == pkg + "@" + ver) {
                return true;
            }
            if (command[3] == "/local/npm-" + pkg + "@" + ver + ".tgz") {
                return true;
            }
            if (command[3] == "/local/" + pkg + "-" + ver + ".tgz") {
                return true;
            }
        }
    }
    if (command[0] == "python3") {
        if (command[1] == "/usr/local/bin/analyze.py") {
            return true;
        }
    }
    if (e == 'pypi' && command[0] == "node" && command[1] == "--max-old-space-size=4069" && command[2].endsWith("jsii-runtime.js")) {
        return true;
    }
    if (command[0] == "/usr/local/bin/python3") {
        if (command[1] == "-m" && command[2] == "pip") {
            if (command[3] == "install" && command[4] == "--pre") {
                if (command[5] == pkg + '==' + ver) {
                    return true;
                }
            }
            if (command[3] == "--disable-pip-version-check" && command[4] == "wheel" && command[5] == "--no-deps" && command[6] == "-w" && command[8] == "--quiet") {
                return true;
            }
        }
        var setuppyscript = "import io, os, sys, setuptools, tokenize; sys.argv[0] = '";
        if ((command[1] == "-c" && command[2].startsWith(setuppyscript)) || (command[1] == "-u" && command[2] == "-c" && command[3].startsWith(setuppyscript))) {
            return true;
        }
        if (["prepare_metadata_for_build_wheel", "get_requires_for_build_wheel", "build_wheel"].indexOf(command[2]) != -1) {
            return true;
        }
        if (command[2] == "install" && command[3] == "--ignore-installed" && command[4] == "--no-user" && command[5] == "--prefix") {
            return true;
        }
    }
    if (command[0] == "npm" && command[1] == "install") {
        if (command[2] == pkg + "@" + ver) {
            return true;
        }
        if (command[2] == ("/local/npm-" + pkg + "@" + ver + ".tgz")) {
            return true;
        }
        if (command[2] == ("/local/" + pkg + "-" + ver + ".tgz")) {
            return true;
        }
    }
    if (command[0] == "ruby") {
        if (command[1] == "/usr/local/bin/analyze.rb") {
            return true;
        }
    }
    if (command[0] == "gem" && command[1] == "install" && command[2] == "-v" && command[3] == ver && command[4] == pkg) {
        return true;
    }
    if (command[0] == "git" && command[1] == "ls-remote") {
        return true;
    }
    if (command[0] == "php") {
        if (command[1] == "/usr/local/bin/analyze.php") {
            return true;
        }
    }
    if (command[0].startsWith('/app/target/debug/build/')) {
        return true;
    }
    if (command[0].startsWith('/usr/local/rustup/toolchains/') && command[0].endsWith('/bin/rustc')) {
        return true;
    }
    if (command[0].startsWith('/usr/local/rustup/toolchains/') && command[0].endsWith('/bin/cargo') && command[1] == 'build') {
        return true;
    }
    if (command[0] == 'cargo' && command[1] == 'build') {
        return true;
    }
    if (command[0] == 'freebsd-version' && command.length == 1) {
        return true;
    }
    allowedCommands = ['rustc', 'rustfmt', 'cc', 'pkg-config', 'ar', 'as', '../lib/bsc.exe'];
    if (allowedCommands.indexOf(command[0]) != -1) {
        return true;
    }
    allowedPkgs = ['bam-builder', 'libtls-sys', 'ogg-opus', 'scicrypt-he', 'libhmmer-sys', 'near-pool-v01', 'openzwave-sys', 'lseq'];
    if (allowedPkgs.indexOf(pkg) != -1) {
      return true;
    }
    if (command[0] == '/usr/bin/ld' && command[1] == '-plugin') {
        return true;
    }
    if (command[0].startsWith('/usr/lib/gcc/') && command[0].endsWith('/cc1')) {
        return true;
    }
    if (command[0].startsWith('/usr/lib/gcc/') && command[0].endsWith('/cc1plus')) {
        return true;
    }
    if (command[0].startsWith('/usr/lib/gcc/') && command[0].endsWith('/collect2')) {
        return true;
    }
    return false;
""";

CREATE TEMP FUNCTION is_known_bad(pkg STRING, ver STRING, c STRUCT<Environment ARRAY<STRING>, Command ARRAY<STRING>>)
RETURNS BOOL 
LANGUAGE js
AS r"""
    let command = c.Command;
    let singleCmd = command.join(" ");
    if (command[0] == "curl" || command[1] == "curl" || command[2] == "curl") {
        return true;
    }
    return false;
""";


CREATE TEMP FUNCTION extract_command(commands ARRAY<STRUCT<Environment ARRAY<STRING>, Command ARRAY<STRING>>>)
RETURNS ARRAY<STRING>
LANGUAGE js AS r"""
  let cmds = [];
  for (let cmd of commands) {
      cmds.push(cmd.Command.join(" "));
  }

  return cmds;
"""
;

CREATE TEMP FUNCTION extract_1command(command STRUCT<Environment ARRAY<STRING>, Command ARRAY<STRING>>, version STRING)
RETURNS STRING
LANGUAGE js AS r"""
  return command.Command.join(" ").replace(version,"VVVV");
"""
;

CREATE TEMP FUNCTION get_env(command STRUCT<Environment ARRAY<STRING>, Command ARRAY<STRING>>, key STRING)
RETURNS STRING
LANGUAGE js AS r"""
    for (let env of command.Environment) {
        let k = key + "=";
        if (env.startsWith(k)) {
            return env.substr(k.length);
        }
    }
    return "";
""";


WITH commands_by_pkg AS (
    SELECT extract_1command((commands), Package.Version) AS command, Package.Ecosystem AS ecosystem, Package.Name AS name, Package.Version AS version, get_env((commands), "npm_package_name") AS executer_name
    FROM
    `ossf-malware-analysis.packages`.analysis_for_phase(PHASE) AS t,
    t.Result.Commands as commands
    WHERE NOT is_known_good(Package.Name, Package.Version, commands, Package.Ecosystem) AND TIMESTAMP_ADD(CreatedTimestamp, INTERVAL LOOKBACK_DAYS DAY) > CURRENT_TIMESTAMP())
SELECT command, COUNT(1) AS occurance, ARRAY_AGG(CONCAT(name, "@", version, "--", executer_name))
FROM commands_by_pkg
WHERE ecosystem = PKG_ECOSYSTEM
GROUP BY command
ORDER BY occurance, command;
```
