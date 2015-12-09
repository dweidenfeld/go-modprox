# ModProx
ModProx is a modification proxy for html documents. You can simply start it and specify some rules in the config.json file (lying in CWD).

ModProx is written in GO.

## Usage
Just create a config.json file and put it next to the executable and start it by executing

    ./go-modprox

## Configuration
The config.json file could look like this:

```json
{
  /* The port to bind on */
  "port": 8080,
  /* A list of modifications that should be processed */
  "modifications": [
    {
      /* RegEx matching URL to process the modification */
      "urlMatch": "^http:\\/\\/example\\.com\\/.*",
      /* The CSS selector from the element you want to get */
      "selector": "h1",
      /* The element to append the text on */
      "appendTo": "title"
    },
    {
      "urlMatch": "^http:\\/\\/jobs\\.daimler\\.com\\/.*",
      "selector": ".meta-info-data > span",
      /* If you want to get an attributes value instead of the text content */
      "attribute": "class",
      /* If there are multiple matches for the selector, the index describes which one should be used (zero-based) */
      "index": 0,
      /* If the value should be trimmed before attached to "appendTo" or the "wrapper" */
      "trim": true,
      /* A wrapper function to wrap the value */
      "wrapper": "<meta name=\"job_nr\" content=\"%s\"/>",
      "appendTo": "head"
    }
  ]
}
```
