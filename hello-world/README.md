## Local

```
% sam local start-api                                                             [4:54:15]
2019-04-16 04:57:57 Found credentials in shared credentials file: ~/.aws/credentials
2019-04-16 04:57:57 Mounting HelloWorldFunction at http://127.0.0.1:3000/hello [GET]
2019-04-16 04:57:57 You can now browse to the above endpoints to invoke your functions. You do not need to restart/reload SAM CLI while working on your functions, changes will be reflected instantly/automatically. You only need to restart SAM CLI if you update your AWS SAM template
2019-04-16 04:57:57  * Running on http://127.0.0.1:3000/ (Press CTRL+C to quit)
```

```
curl http://127.0.0.1:3000/hello
```
