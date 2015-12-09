# ModProx
ModProx is a modification proxy for html documents. You can simply start it and specify some rules in the config.json file (lying in CWD).

ModProx is written in GO.

## Usage
Just create a config.json file and put it next to the executable and start it by executing

    ./go-modprox

## Configuration
The config.json file could look like this:

### Options
| Key                      | Description                                                                                               |
|--------------------------|-----------------------------------------------------------------------------------------------------------|
| port                     | The port to bind on                                                                                       |
| modifications            | A list of modifications that should be processed                                                          |
| modification[].urlMatch  | RegEx matching URL to process the modification                                                            |
| modification[].selector  | The CSS selector from the element you want to get                                                         |
| modification[].appendTo  | The element to append the text on                                                                         |
| modification[].attribute | If you want to get an attributes value instead of the text content                                        |
| modification[].index     | If there are multiple matches for the selector, the index describes which one should be used (zero-based) |
| modification[].trim      | If the value should be trimmed before attached to "appendTo" or the "wrapper"                             |
| modification[].wrapper   | A wrapper function to wrap the value                                                                      |

### Example
```json
{
  "port": 8080,
  "modifications": [
    {
      "urlMatch": "^http:\\/\\/example\\.com\\/.*",
      "selector": "h1",
      "appendTo": "title"
    },
    {
      "urlMatch": "^http:\\/\\/example\\.com\\/.*",
      "selector": ".meta-info-data > span",
      "attribute": "class",
      "index": 2,
      "trim": true,
      "wrapper": "<meta name=\"id\" content=\"%s\"/>",
      "appendTo": "head"
    }
  ]
}
```

# License
The MIT License (MIT)

Copyright (c) 2015 Dominik Weidenfeld

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
