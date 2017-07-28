
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
