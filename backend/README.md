# Backend

This directory contains the backend.

Right now the backend is just a static server for the front end.

# Developer Guide

## Test Data

Google Documents can be downloaded using the CLI and saved to create
data to be used in tests; e.g.

```
./build/bin/server getdoc \
    --credentials-file=${CREDENTIALS_FILE} \
    --doc=${GOOGLE_DOC_ID} \
    --format=text \
    -o ${PATH_TO_SAVE_FILE}
```
The output of the Google Cloud Natural Language API can be downloaded as follows:

```
./build/bin/server getentities \
    --input ${INPUT_TEXT_FILE} \
    -o ${PATH_TO_SAVE_OUTPUT}
```

The input text file should be a string of linear text. This can be obtained from
a Google Document using the previous command.