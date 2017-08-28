# timeular-zei-linux
Very basic Linux client for the Timeular ZEI, due to lack of a Linux implementation for the Timeular Zei

# Learn more
 - [The timeular ZEI](https://timeular.com/product/zei/)
 - [Timeular public API](http://developers.timeular.com/public-api/)

# Usage

1. Create an API key on your [profile page](https://profile.timeular.com)
2. Copy the API key and secret into a config.json file
```json
{
  "apiKey": "my-api-key",
  "apiSecret": "my-api-secret"
}
```
3. Go get
4. Go build
5. Run the binary as superuser (sudo ./timeular-zei-linux)

It is necessary to run as superuser in order to communicate with the Bluetooth device