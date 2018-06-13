
0.8.0 / 2018-06-13
==================

BACKWARDS INCOMPATIBILITIES:

  * Remove CORS configuration from subscription (#458)
  * Refactored Subscription model (#429)
  * Remove invoke functionality (#432)
  * HTTP CloudEvent coercion (#421)
  * Move Promethues metrics endpoint under /v1/ prefix. Closes #423 (#426)
  * Handle +json Content-Type in accordance with RFC6839 (#416)
  * Use Content-Type application/cloudevents+json for http provider (#415)

IMPROVEMENTS:

  * Add CORS Config API (#460)
  * Update docs about event types, events api authorization (#454)
  * Switch to AtomicPut in Create methods (#457)
  * Include authorization result in the event extensions (#456)
  * Add endpoint for updating event type (#455)
  * Event Type Authorizer (#441)
  * Event types (#433)
  * Add Links to ReadMe (#430)
  * Update Prometheus docs
  * Cleanup README and documentation. Add docs for Prometheus Metrics. (#422)
  * Move metric definition to the same file as other metrics (#424)
  * Update examples (#420)
  * Update README.md (#419)

BUG FIXES:

  * Fix payload normalization in http.request events. Closes #434 (#459)
  * Remove duplicated links in the README (#431)
  * Disable TravisCI notifications (#417)


0.7.0 / 2018-04-26
==================

BACKWARDS INCOMPATIBILITIES:

  * require CloudEvents Source be a URI (#414)
  * cloudEvents integration (#404)
  * make sure event body is a string (#393)
  * cleanup error messages and provide consistency in error reporting (#396)
  * remove unused providers (emulator, weighted) (#387)
  * refactor Prometheus metrics. Less metrics, more labels (#384)
  * use different messages on lambda errors (#381)
  * add more detailed metrics (#379)
  * improve function invocation error (#377)

IMPROVEMENTS:

  * update README.md
  * update README.md
  * update readme (#406)
  * add Reliablity Guarantees section in README (#403)
  * add UpdateSubscription endpoint (#398)
  * add SQSProvider (#399)
  * add missing httpapi tests (#401)
  * add AWS Firehose provider (#394)
  * improve Provider tests (#397)
  * add OpenAPI specification. (#395)
  * add Kinesis provider. (#392)
  * add codecov integration. (#391)
  * introduce Providers concept in the codebase. (#389)
  * hide sensitive information in logs (#385)
  * minimalize Dockerfile (#378)

BUG FIXES:

  * fix minor issue with open api spec (#413)
  * fix handling of mime type (#405)
  * fix installation script (update artifact format). Closes #409 (#412)
  * update Docker image in running-locally.md (#386)
  * fix for UpdateFunction (#382)
  * update create subscription docs regarding path param
  * update README with info about Docker image. Closes #325 (#375)

0.6.0 / 2018-02-19
==================

BACKWARDS INCOMPATIBILITIES:

  * flatten header array so it's easier to access headers (#374)
  * publish binaries as zip files (#371)
  * remove checking space in update function payload (#367)
  * add space as a first-class citizen in Config API (#365)
  * fix listing all subscriptions (#366)
  * add support for spaces (#362)
  * standardize errors returned by both Events and Config API (#359)
  * add subscription support for invoke event (#355)

IMPROVEMENTS:

  * prevent from removing function with subscriptions. Closes #208 (#370)
  * replace FDK examples with SDK. Closes #363 (#368)
  * update HTTP response object docs
  * refactor packages structure to avoid cyclic dependecies (#358)
  * update deps
  * change behaviour of hosted domain (#353)
  * improve support for hosted domains (#351)
  * add flag for configuring backlog lenght (#349)
  * add flag for configuring number of workers (#348)
  * make metrics name consistent
  * add more metrics about processing sync/async events
  * remove old comments
  * refactor prometheus metrics (#347)
  * increase read/write timeout on events API
  * add ssl certs to Docker image
  * add support for platform subdomains (#342)
  * expose used ports in Dockerfile
  * add session token support in AWS creds. Closes #329 (#339)
  * add host field in HTTP event. Closes #327 (#338)
  * update libkv & etcd packages (#335)
  * add CORS support. Closes #328, closes #309. (#334)
  * add docs about developing EG locally
  * Plugin System (#330). Closes #147
  * add tests for internal/kv package
  * add Path parameter support for async subscriptions (#326)
  * add support for parameters in HTTP subscription path (#322). Closes #217.
  * add unit tests for functions service (#323)
  * update HTTP event docs. Closes #275
  * move event to separate package. Cleanup topic/event type vocabulary (#320)
  * remove obsolete if statement. It's no longer valid in etcd3
  * add support for coveralls (#318)
  * bump serverless/libkv fork
  * switch to libkv fork with etcd v3 support. It fixes race condition in test and improve general stability. Closes #222 (#317)
  * add dockerignore file to speed up docker build. Closes #314
  * update dep installing in Dockerfile
  * update slack link
  * switch to dep (#313)
  * remove slack badge
  * bring back cache debug logs (#312)
  * add support for HTTP response object. Closes #245 (#291)
  * add info about versioning (#304)
  * prefix log statements with timestamp. Closes #251 (#300)
  * add slack link (#303)
  * use docker multi-stage build to avoid compile time dependencies (#301)
  * GitHub templates and docs improvements (#302)
  * add example app

BUG FIXES:

  * fix typo (#373)
  * fix status code when creating subscription (#372)
  * run goveralls only for PR build (#364)
  * fix failing router tests (#360)
  * rename function property (#352)
  * format README
  * fix exposing detailed AWS SDK error by event API. Closes #344 (#350)
  * fix interface mismatch for plugins, exclude hashicorp packages. Closes #345 (#346)
  * fix type not registered in gob for plugin system
  * fix extracting path from domain (#343)
  * fix type in README.md (#341)
  * fix typo in README (#336)
  * fix typo about subscription removal (#332)
  * fix type in error message (#331)
  * update link to Slack in issue template
  * cleanup confusion in clustering paragraph (#324)
  * fix conflict in subscriptions ID. replace - with , as a subscription ID separator. Closes #170 (#321)
  * fix issues reported by https://goreportcard.com/report/github.com/serverless/event-gateway
  * fix rc when starting APIs. Closes #310 (#311)
  * add docs for using EG with Docker (#307)
  * fix framework links (#308)
  * fix framework link (#305)
  * fix meetups link
  * fix readme.md typo (#297)

0.5.15 / 2017-08-17
===================

  * allow path without trailing / and normalize method. Closes #272 (#296)
  * Update README.md
  * add webbhook to function types
  * typo in README
  * remove unused asset
  * update quickstart section (#294)
  * README improvements (#293)
  * fix installation script
  * readme improvements (#292)

0.5.14 / 2017-08-15
===================

  * temporary remote HTTP response support

0.5.13 / 2017-08-15
===================

  * update emulator's invoke endpoint (#288)
  * change internal function error name (#287)
  * add support for HTTP response object. Closes #245 (#286)
  * run tests before lint on TravisCI

0.5.12 / 2017-08-14
===================

  * improve logging. Add missing logs. Better error reporting. Closes #265 (#285)
  * add tests for subscription package
  * rename pubsub package to subscriptions

0.5.11 / 2017-08-14
===================

  * add path and method as a part of http event. Closes #273 (#284)
  * fix log messages and add missing logs. Closes #280. Closes #277. (#283)

0.5.10 / 2017-08-13
===================

  * decode invoke payload into empty interface (#281)i

0.5.9 / 2017-08-13
==================

  * add Emulator Provider (#279)

0.5.8 / 2017-08-11
==================

  * add emitting internal gateway event (#274)

0.5.7 / 2017-08-11
==================

  * fix HTTP event structure
  * standarize error names
  * rename ErrorMalformedJSON to ErrMalformedJSON

0.5.6 / 2017-08-11
==================

  * allow all headers for CORS. Closes #269 (#271)

0.5.5 / 2017-08-11
==================

  * return 404 if backing function not found for invoking or http req. Closes #265 (#268)
  * fix HTTP schema field. Rename it to body. Closes #266 (#267)
  * fix comment
  * refactor integration test. Move them to router package.
  * refactor targetcache package. Move it to internal package.
  * refactor metrics package. Move it to internal package.
  * refactor util package. Move it to internal package.
  * refactor db package. Move it to internal package.
  * cleanup docs folder
  * remove outdated usecases doc

0.5.4 / 2017-08-10
==================

  * return HTTP error when JSON payload is malformed. Closes #241 (#261)
  * add logging about address on which APIs are listening. Closes #246 (#260)
  * remove redundant logs. Log HTTP event payload. (#259)
  * output version to stdout

0.5.3 / 2017-08-10
==================

  * add cors headers. Closes #255 (#257)
  * add -log-format option. Closes #247 (#256)

0.5.2 / 2017-08-09
==================

  * add log for successful function invocation. Closes #254
  * allow dots, dash, underscore in event and function name. Closes #250 (#253)
  * fix sending header twice. Closes #249 (#252)
  * cleanup router logging

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

