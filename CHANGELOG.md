
0.5.1 / 2017-08-08
==================

  * update FDK examples
  * update examples; add FDK examples; add like to FDK. Closes #224. Closes #223. Closes #205 (#236)
  * minor logging & httpapi fixes. Closes #230. Closes #228. (#235)
  * return 204 (No Content) for successful subscription delete. Closes #233 (#234)
  * update log level for functions and pubsub packages

0.5.0 / 2017-08-04
==================

  * remove secured env vars from travis.yml as they don't work for public repos
  * update GITHUB_TOKEN
  * add support for HTTP event schema. Closes #201. Closes #203 (#212)
  * fix HTTP API method for updating function. Closes #204 (#211)
  * fix a spelling typo (#209)
  * fix built package in Dockerfile. Closes #210
  * fix lint errors
  * remove time from dev logs. Closes #167
  * cleanup emitted logs. Closes #167
  * add more Info level logs about emitted event and config API actions. Closes #167 (#202)
  * use existing content-type mime values. Closes #199 (#200)
  * refactor `httplisteners` package. Move it to `api` package.
  * implement event schema and add support for content-type header. Closes #161. Closes #145 (#198)
  * update community links (#196)
  * move main package to cmd subdir
  * switch to public Travis

0.4.1 / 2017-07-31
==================

  * fix header name

0.4.0 / 2017-07-31
==================

  * add API for sync invocation. Closes #183 (#192)

0.3.0 / 2017-07-28
==================

  * add missing package with integration tests
  * fix package name
  * move httpapi to util
  * refactor tests. Clearer structure
  * add endpoint for updating function. Closes #190 (#191)
  * update AWS Lambda func readme docs. Refactor calling Lambda func.
  * allows exec via all AWS configs. Closes #186 (#187)
  * Remove unsupported use-cases. Add community sections to readme. (#188)
  * add status endpoint. Closes #177 (#185)
  * fix emitting custom events. Closes #160 (#184)
  * fix readme inconsistency about the emit functionality (#182)
  * update default ports and cli params for setting ports. Closes #180 (#181)
  * Update examples (#178)
  * Add license file (#176)
  * fix cyclomatic complexity of validation function. #174
  * validate Functions Disc API. Closes #174 (#175)
  * add info about releases to README

0.2.0 / 2017-07-26
==================

  * prevent from registering function with the same name. Closes #173
  * add changelog file
  * update Function Discovery API (provider field). Closes #158 (#172)
  * add goreleaser for publishing binaries. Closes #164 (#171)
  * add docs about getting all functions
  * add endpoint for listing registered functions. Closes #159 (#166)
  * fix default etcd ports (#165)
  * update functions discovery docs. Closes #122 (#157)
  * add events api description. Closes #152 (#156)
  * fix api port in the readme
  * minor readme updates
  * add validation for HTTP subscription. Closes #143 (#144)
  * fix creating http subscriptions for the same path and method. Closes #141 (#142)
  * cleanup README. Closes #125 (#140)
  * fix TravisCI build. Replace curl from GH with APT (#139)
  * add quick start section in README (#138)
  * remove endpoints HTTP API. Closes #119 (#133)
  * add Makefile with build target (#137)
  * fix Travis CI build (#136)
  * remove deplyoment info. Closes #134 (#135)
  * remove topics HTTP API. Closes #120 (#130)
  * add missing function discovery method (#131)
  * remove publisher and enable emitting events. Closes #121. Closes #129. (#128)
  * add use cases (#127)
  * start workers (#126)
  * add info about default config api port
  * update API path
