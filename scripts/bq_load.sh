#!/bin/bash

if [ -z "$PROJECT_ID" ]; then
    echo "PROJECT_ID must be set"
    exit 1
fi

if [ -z "$LOAD_DATASET" ]; then
    echo "LOAD_DATASET must be set"
    exit 1
fi

if [ -z "$LOAD_TABLE_PREFIX" ]; then
    echo "LOAD_TABLE_PREFIX must be set"
    exit 1
fi

if [ -z "$DEST_DATASET" ]; then
    echo "DEST_DATASET must be set"
    exit 1
fi

if [ -z "$DEST_TABLE" ]; then
    echo "DEST_TABLE must be set"
    exit 1
fi

if [ -z "$RESULT_BUCKET" ]; then
    echo "RESULT_BUCKET must be set"
    exit 1
fi

if [ -z "$SCHEMA_FILE" ]; then
    echo "SCHEMA_FILE must be set"
    exit 1
fi

union=""

for bucket_prefix in `gsutil ls "$RESULT_BUCKET"`; do
    prefix=`echo "$bucket_prefix" | sed "s|$RESULT_BUCKET/\([^\]*\)/|\1|g"`
    clean_prefix=`echo "$prefix" | tr -c -d "[:alnum:]"`
    table_name="$LOAD_TABLE_PREFIX$clean_prefix"

    echo "## Loading $bucket_prefix into \`$PROJECT_ID.$LOAD_DATASET.$table_name\`."
    bq load \
        --headless \
        --project_id="$PROJECT_ID" \
        --dataset_id="$LOAD_DATASET" \
        --replace \
        --time_partitioning_type="DAY" \
        --time_partitioning_field="CreatedTimestamp" \
        --source_format="NEWLINE_DELIMITED_JSON" \
        --max_bad_records=10000 \
        "$table_name" "$bucket_prefix*"  "$SCHEMA_FILE"

    if [ $? -eq 0 ]; then
        echo "Loading \`$PROJECT_ID.$LOAD_DATASET.$table_name\` done."
    else
        echo "Loading \`$PROJECT_ID.$LOAD_DATASET.$table_name\` failed. Aborting"
        exit 1
    fi

    # Construct a UNION query for joining the prefix shards together
    subquery="SELECT * FROM \`$PROJECT_ID.$LOAD_DATASET.$table_name\`"
    if [ -n "$union" ]; then
      union="$union UNION ALL "
    fi
    union="$union$subquery"
done

# Query to check the destination table schema has not changed.
schema_check_query="
WITH unique_schemas AS (SELECT
 TO_JSON_STRING(
    ARRAY_AGG(STRUCT(
      IF(is_nullable = 'YES', 'NULLABLE', 'REQUIRED') AS mode, column_name AS name, data_type AS type)
    ORDER BY ordinal_position), TRUE) AS schema
FROM \`$PROJECT_ID.$DEST_DATASET\`.INFORMATION_SCHEMA.COLUMNS
WHERE table_name = \"$DEST_TABLE\"
UNION DISTINCT
SELECT
 TO_JSON_STRING(
    ARRAY_AGG(STRUCT(
      IF(is_nullable = 'YES', 'NULLABLE', 'REQUIRED') AS mode, column_name AS name, data_type AS type)
    ORDER BY ordinal_position), TRUE) AS schema
FROM \`$PROJECT_ID.$LOAD_DATASET\`.INFORMATION_SCHEMA.COLUMNS
WHERE table_name = \"$table_name\")
SELECT IF(c = 1, \"OK\", ERROR(\"Schemas have changed.\")) FROM (SELECT COUNT(*) AS c FROM unique_schemas);"

echo "## Checking \`$PROJECT_ID.$DEST_DATASET.$DEST_TABLE\` schema."
echo "Executing query: '$schema_check_query'"

bq query --headless --nouse_legacy_sql --project_id="$PROJECT_ID" "$schema_check_query"

# Ensure we abort if the exit code was not successful.
if [ $? -eq 0 ]; then
    echo "Schemas match. Continuing."
else
    echo "Schemas do not match. Aborting."
    exit 1
fi

# Query to populate the destination table.
#
# If the table exists it will be replaced.
replace_query="CREATE OR REPLACE TABLE \`$PROJECT_ID.$DEST_DATASET.$DEST_TABLE\` LIKE \`$PROJECT_ID.$LOAD_DATASET.$table_name\` PARTITION BY TIMESTAMP_TRUNC(CreatedTimestamp, DAY) OPTIONS(expiration_timestamp=NULL) AS $union;"

echo "## Updating \`$PROJECT_ID.$DEST_DATASET.$DEST_TABLE\` from shards."
echo "Executing query: '$replace_query'"

bq query --headless --nouse_legacy_sql --project_id="$PROJECT_ID" "$replace_query"
