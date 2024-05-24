
# AutoLoadDotEnv Algorithm

## Overview

The `AutoLoadDotEnv` algorithm is designed to automatically load environment variables from `.env` files located in a directory and its parent directories. It offers flexibility and control over how environment variables are loaded into your application.

## Features

### Recursive Search

The algorithm starts its search for `.env` files from a specified directory and continues recursively up to its parent directories until either a maximum depth is reached or it reaches the root directory.

### Environment Variable Control

- `DOTENV_DIR`: Specifies the starting point for the recursive search for `.env` files.
- `DOTENV_FILE`: If set, targets a single `.env` file and short-circuits the full search. Only the variables in this file will be loaded.

## Algorithm Steps

1. Read `DOTENV_DIR` to determine the starting directory for the search. If not set, use the current working directory.
2. If `DOTENV_FILE` is set, read only that file and skip the full search.
3. Start the recursive search from the starting directory:
    - Look for a `.env` file in the current directory.
    - If found, load the environment variables from this file.
    - Move to the parent directory and repeat the steps.

### Precedence Rules

The algorithm respects a precedence rule when multiple `.env` files are found during the recursive search. Specifically, environment variables set by `.env` files in parent directories will overwrite those set by `.env` files in child directories.

For example, if both `level1` and `level2` directories contain `.env` files, and `level1` is a parent of `level2`, the variables in `level1/.env` will take precedence over those in `level2/.env`.

## Error Handling and Debugging

If a `.env` file is found but cannot be parsed, a debug log will be written using `slog.Default` logger, and the algorithm will continue its search up the directory tree.
