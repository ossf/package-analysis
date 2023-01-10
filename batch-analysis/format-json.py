#!/usr/bin/python3

"""Custom tool to pretty-print JSON with certain fields compacted

Adapted from source of `python -m json.tool`
reference: github.com/python/cpython/blob/main/Lib/json/tool.py

"""

import argparse
import json
import pathlib
import re
import sys


# Changes JSON structs that are formatted like:
#     {
#         "Name": "...",
#         "Type": "..."
#     }
# into ones like
#     { "Name": "...", "Type": "..." }
name_type_substitution = (
    re.compile('{$\\n^\\s*"Name": ?"(.*)",$\\n^\\s*"Type": ?"(.*)"$\\n^\\s*}', re.MULTILINE),
    '{ "Name": "\\1", "Type": "\\2" }'
)

# Changes JSON structs that are formatted like:
#     {
#         "Value": ..., (may not be a string)
#         "Raw": "..."
#     }
# into ones like
#     { "Value": ..., "Raw": "..." }
#
value_raw_substitution = (
    re.compile('{$\\n^\\s*"Value": ?(.*),$\\n^\\s*"Raw": ?"(.*)"$\\n^\\s*}', re.MULTILINE),
    '{ "Value": \\1, "Raw": "\\2" }'
)


# Reformats a JSON string to apply the substitutions above,
# while maintaining indent level
def reformat_json(json_string: str) -> str:
    sub1 = re.sub(*name_type_substitution, json_string)
    sub2 = re.sub(*value_raw_substitution, sub1)
    return sub2


def make_arg_parser() -> argparse.ArgumentParser:
    prog = 'format_json.py'
    description = 'Pretty-prints JSON data from analysis output'
    parser = argparse.ArgumentParser(prog=prog, description=description)

    parser.add_argument('infile', nargs='?', default=sys.stdin,
        type=argparse.FileType(encoding="utf-8"), help='input JSON file')

    parser.add_argument('outfile', nargs='?', default=None,
        # type=Path means that the file is not automatically opened and truncated,
        # so in-place formatting is possible using the same input and output file
        type=pathlib.Path, help='output (formatted) JSON file')

    return parser


def main():
    arg_parser = make_arg_parser()
    options = arg_parser.parse_args()

    with options.infile as infile:
        # pretty print with newlines and indent with 4 spaces,
        pretty_printed_json = json.dumps(json.load(infile), indent=4)

    custom_formatted_json = reformat_json(pretty_printed_json)

    if options.outfile is None:
        out = sys.stdout
    else:
        out = options.outfile.open('w', encoding='utf-8')
    with out as outfile:
        outfile.write(custom_formatted_json)
        outfile.write('\n')


if __name__ == '__main__':
    try:
        main()
    except BrokenPipeError as exc:
        sys.exit(exc.errno)
    except ValueError as e:
        raise SystemExit(e)

