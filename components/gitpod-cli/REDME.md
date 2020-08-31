# The Gitpod CLI

## What is it

This is the source code for the Gitpod CLI or `gp` this is made to streamline the Gitpod experience

## Security

### Problem

We cannot allow `gp` to access services directly. As this could in turn allow a user to directly interface with a service potentially giving access to the other services.

### Solution

Instead `gp` reaches into the Theia frontend which in-turn reaches into a Gitpod service. 

## How to access other services through Theia frontend.

An example of this is in the `gp open` command whose source code is in `cmd/open.go` First to access  Theia frontend we use `service, err := theialib.NewServiceFromEnv()` which allows us to access the Theia service located in `components/theia/packages/gitpod-extension/src/browser`. Now first it needs to implement an method that takes in a request to Theia and returns a response. To do this we go to `pkg/theialib/protocol.go` And define 3 things: A `struct` that serves as the request containing all the data that Theia will need to complete said request, in this case

```go
type OpenFileRequest struct {
 	Path string `json:"path"`
}
```

Here the only data we are passing is the path to the file we want Theia to open. notice the annotation this is important for later when the Theia portion is implemented.


Next we define a response this basically follows the same pattern we create a struct with all the data we want to receive in this case the struct is empty

```go
type OpenFileResponse struct{}
```

Finally on the request side of things we create a method on the interface `TheiaCLIService` that will be implemented on the Theia side. You can name this anything but in this case it is
```go
OpenFile(OpenFileRequest) (*OpenFileResponse, error)
```

Notice that it takes in the request struct and returns a pointer to the response along with an error.

The last step is to run the `generate-theia-protocol.go` script which should generate this on the Theia side.

We can then technically use these APIs in our code base (they won't be able to do anything as of yet).

Next, Let's implement these APIs in Theia

First let's implement the client. The file is located at `components/theia/packages/gitpod-extension/src/browser/cli-service-client.ts ` Make sure you import the request and response from `../common/cli-service` Implement a function in the `CliServiceClientImpl` class which takes in the request and returns a Promise holding the response.

In this case that function looks like

```ts
async closeFile(params: CloseFileRequest): Promise<CloseFileResponse> {
  await this.applicationStateService.reachedState("ready");
  const path = params.path;
  const uri = new URI(`file://${path}`);
  this.editorManager.getByUri(uri).then(e => {e?.close()});
  return {}
}
```

Finally implement the server located at `components/theia/packages/gitpod-extension/src/node/cli-service-server.ts` Implement the logic once again taking in the request and returning a promise. First off make sure the client has connected in the first line using `await this.firstClientConnected.promise` then operate on the client calling the methods in this case the full method looks like

```ts
async openFile(params: OpenFileRequest): Promise<OpenFileResponse> {
    await this.firstClientConnected.promise;
    await Promise.all(this.clients.map(c => c.openFile(params)));
    return {};
}
