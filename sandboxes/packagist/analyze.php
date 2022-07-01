#!/usr/bin/env php
<?php

const PHP_EXTENSION = "php";

class Package {
    public string $name = "";
    public ?string $version = NULL;
    public ?string $local_path = NULL;

    public function __construct(string $name, ?string $version = NULL, ?string $local_path = NULL) {
        $this->name = $name;
        $this->version = $version;
        $this->local_path = $local_path;
    }

    public function artifactPath(): ?string {
        if ($this->local_path === NULL) {
            return NULL;
        }
        $path = realpath(dirname($this->local_path));
        if (!$path) {
            throw new Exception("Unknown local file path");
        }
        return $path;
    }

    public function packageVersion(): string {
        if ($this->version != NULL) {
            return "{$this->name}:{$this->version}";
        } else {
            return $this->name;
        }
    }

    public function packageBasePath(): string {
        $path = realpath("vendor" . DIRECTORY_SEPARATOR . $this->name);
        if (!$path) {
            throw new Exception("Could not find package base path");
        }
        return $path;
    }
}

function walkDir(string $dir): array {
    $results = array();
    $files = scandir($dir);
    foreach ($files as $value) {
        $path = realpath($dir . DIRECTORY_SEPARATOR . $value);
        if (!is_dir($path)) {
            $results[] = $path;
        } else if ($value != "." && $value != "..") {
            $results = array_merge($results, walkDir($path));
        }
    }
    return $results;
}

function makeCmd(string $cmd, string ...$args): string {
    $safecmd = escapeshellcmd($cmd);
    foreach ($args as $arg) {
        $safecmd .= " " . escapeshellarg($arg);
    }
    return $safecmd;
}

function install($package) {
    $args = array(
        "init",
        "-n", // no interaction
        "--name=app/app",
    );
    $artifactPath = $package->artifactPath();
    if ($artifactPath !== NULL) {
        $arg = array("type" => "artifact", "url" => $artifactPath);
        $repositoryArg = sprintf('--repository=%s', json_encode($arg));
        $args[] = $repositoryArg;
    }
    $retval = 0;
    passthru(makeCmd("composer.phar", ...$args), $retval);
    if ($retval != 0) {
        throw new Exception("Failed to setup artifact repository");
    }
    passthru(makeCmd("composer.phar", "require", "--no-progress", $package->packageVersion()), $retval);
    if ($retval === 0) {
        print("Install succeeded.\n");
    } else {
        print("Install failed.\n");
        exit(1);
    }
}

function import($package) {
    require 'vendor/autoload.php';

    $files = walkDir($package->packageBasePath());
    foreach ($files as $file) {
        if (pathinfo($file, PATHINFO_EXTENSION) === PHP_EXTENSION) {
            try {
                print("Importing $file\n");
                include_once($file);
            } catch (Throwable $t) {
                print("Failed to import $file\n");
                print($t->getTraceAsString());
            }
        }
    }
}

const PHASES = array(
    "all" => array("install", "import"),
    "install" => array("install"),
    "import" => array("import"),
);

$args = $argv;
$script = array_shift($args);

$arg_count = count($args);
if ($arg_count < 2 || $arg_count > 4) {
    fwrite(STDERR, "Usage: $script [--local file | --version version] phase package_name\n");
    exit(1);
}

# Parse the arguments manually to avoid introducing unnecessary dependencies
# and side effects that add noise to the strace output.
$local_path = NULL;
$version = NULL;
switch($args[0]) {
    case '--local':
        array_shift($args);
        $local_path = array_shift($args);
        break;
    case '--version':
        array_shift($args);
        $version = array_shift($args);
        break;
}

$phase_name = array_shift($args);
$package_name = array_shift($args);

if (!array_key_exists($phase_name, PHASES)) {
    fwrite(STDERR, "Unknown phase $phase_name specified.\n");
    exit(1);
}

$package = new Package($package_name, $version, $local_path);

foreach (PHASES[$phase_name] as $phase) {
    call_user_func($phase, $package);
}