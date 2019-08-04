healthy
=======

![build status](https://travis-ci.org/roman-mazur/healthy.svg?branch=master)

A very simple health checker for your site. The root package contains a library that can be used in your Go program.

`healthyd` package contains a command that reads JSON config and start a deamon running configured checks.
May be run on rpi.

Installation
------------

**With go get**

```
go get rmazur.io/healthy/healthyd
```

Usage
-----
Command line options:
```
  -config-file string
    	configuration file path
```

For example
```
healthyd -config-file my-checks.json
```

Configuration file example.
```json
{
  "twillio": {
    "accountId": "<myAccountId>",
    "authToken": "<myAuthToken>>",
    "from": "<myTwillioNumber>",
    "to": "<myPersonalNumber>"
  },
  "reportFailuresCount": 3,
  "firstRetryDelay": "3s",
  "httpChecks": [
    {
      "url": "https://example.xom",
      "expectedStatusCode": 200,
      "timeout": "2s",
      "period": "5m",
      "flex": "1m"
    },
    {
      "url": "https://google.com",
      "expectedStatusCode": 200,
      "timeout": "2s",
      "period": "10m",
      "flex": "2m"
    }
  ]
}
```
Here we define 2 HTTP checks run with different time intervals (`period` +/- `flex`).
If `twillio` config is defined, the daemon will send invoke Twillio APIs to send SMS notifications.

License
-------
    Copyright 2019 Roman Mazur
    
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
    
       http://www.apache.org/licenses/LICENSE-2.0
    
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
