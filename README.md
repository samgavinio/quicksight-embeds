### Quicksight Embedding

:warning: The code in this repository is for exploration purpose, not production ready and is likely to have defects.

This is a demo application of how a Quicksight Dashboard can be embedded in a Go [(Echo)](https://github.com/labstack/echo) application. [Amazon Cognito](https://docs.aws.amazon.com/cognito/latest/developerguide/what-is-amazon-cognito.html) is used as a user pool.

Reference: https://aws.amazon.com/blogs/big-data/embed-interactive-dashboards-in-your-application-with-amazon-quicksight/

#### Pre-requisites

* The web application must be hosted as a publicly accessible HTTPS application and the hostname must be correctly whitelisted in Quicksight. If running locally, consider tools like [ngrok.io](https://ngrok.com/) to expose your service with a public URL.
* The Cognito user pool must be configured already and the relevant users must already exist in the user pool. The app presently does not support new registrations.
* A Quicksight Group must already exist and must be given the appropriate permission to access the relevant dashboard.
* An IAM Role that will allow invoking the appropriate quicksight permission must already exist. Refer to the [documentation](https://docs.aws.amazon.com/quicksight/latest/user/embedded-dashboards-with-iam-setup-step-2.html)

### How to build

#### Build the frontend components.

```
npm install
npm run build
```

#### Building the backend Go components:

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies. Running either of the following commands should pull-in all the needed dependencies.

```
go build
// or
go test
```

#### Configure the environment variables

```
cp config.yml.example config.yml

// modify the environment varaibles as needed
SESSION_KEY: "1234567890ABCDEFGHIJKLMNOPQRSTUV"
AWS_ACCOUNT_ID: "1111222333"
AWS_REGION: "ap-southeast-1"
COGNITO_CLIENT_ID: "xxxxxxxxxxxxxxxxx"
COGNITO_USER_POOL_ID: "ap-southeast-1_xxxxxxxx"
QUICKSIGHT_IAM_ROLE_NAME: "QuickSightReaderRole"
QUICKSIGHT_GROUP_NAME: "embed-readers"
QUICKSIGHT_DASHBOARD_ID: "xxxxx-xxx-xxxx-xxxx-xxxxxxx"
```

#### Run the application

```
// in development
go run main.go

// access the app by going to [hostname]:1323 in your browser of choice.
```

*Note: Pass-in a set of AWS credentials to the application that will be used to call STS assume role for the Quicksight reader role. If running on AWS, the recommended approach is to attach an IAM Role on your compute resource. Otherwise you may pass in your credentials as environment variables directly as described in the [documentation](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials).*


### License

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.