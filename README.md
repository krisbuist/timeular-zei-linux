# timeular-zei-linux
Very basic Linux client for the Timeular ZEI, due to lack of a Linux implementation for the Timeular Zei

# Learn more
 - [The timeular ZEI](https://timeular.com/product/zei/)
 - [Timeular public API](http://developers.timeular.com/public-api/)

# Usage

1. Create an API key on your [profile page](https://profile.timeular.com)
2. Copy the API key and secret into the `config.json` file in the root of the project
    ```json
    {
      "apiKey": "my-api-key",
      "apiSecret": "my-api-secret"
    }
    ```
    1. Add the serial number of ZEI device you want to use with this computer to `config.json` as well if you use *more than one* ZEI device (e.g. one at work and one at home)
        ```json
        {
          "apiKey": "my-api-key",
          "apiSecret": "my-api-secret",
          "deviceSerial": "TZ******"
        }
        ```
3. Run `go get`
4. Run `go build`
5. Give the client enough capabilities to open RAW sockets: `sudo setcap cap_net_raw,cap_net_admin+eip timeular-zei-linux`
6. Start the application: `./timeular-zei-linux`
