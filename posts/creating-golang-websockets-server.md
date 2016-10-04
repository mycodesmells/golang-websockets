# Creating Websockets Server in Go

Long gone are times when all the communication between web application server and its clients was limited to request-response communication. Nowadays, especially in systems where it is important to notify users about various events, it's absolutely necessary to use websockets in order to push information from the central point. Let's take a look on how to create simple WS server using Golang.

### Standard package

With other languages, like Java Script, you generally start creating a WS server by looking for the best 3rd party library that can make this process as easy as possible. In Golang world, though, the approach is completely different - you should stick to the core language packages so that often you need to create most (or even whole) application code yourself. Fortunately, `golang.org/x/net/websocket` is absolutely sufficient for a simple WS server (and probably for much more) - let's dive in and stick to it!

### Idea of the solution

In our simple example application we would like to accomplish three things:
- accept client's connection and send a greeting message,
- broadcast clients' messages to all active connections,
- create an endpoint so that server can push custom message to all clients.

From the technical standpoint, keeping a two-way communication with each client requires us to have two _loops_:
- one waits for incoming client's message and sends it to the server to broadcast it,
- the other waits for broadcasted messages to direct them to the client.

### Working solution

**Entry point:** We start by defining a structure for our message. Since we'd like to use JSON as transport format, we need to add appropriate encodings:

    type Message struct {
        Author string `json:"author"`
        Body   string `json:"body"`
    }

Then we'll need two endpoints - one will handle whole WS logic, while the other will read user's input (given as path parameter) and send it to all clients:

    ...
    func main() {
        http.HandleFunc("/broadcast/", broadcastHandler)
        http.Handle("/ws", wsHandler)

        http.ListenAndServe(":3000", nil)
    }
    ...

**Client:** Next, we need to create some representation of a client. It will need three properties: WS connection, channel for incoming messages and channel for notifying that the connection should be terminated:

    ...
    type Client struct {
        connection *websocket.Conn
        ch         chan *Message
        close      chan bool
    }
    ...

Listening for incoming and outgoing messages is just a simple `select` statement which waits for one of the events to determine what action should be performed. As both of those listenings should never end, we need to wrap them with `for{ ... }`:

    ...
    func (c *Client) listenToWrite() {
        for {
            select {
            case msg := <-c.ch:
                log.Println("Send:", msg)
                websocket.JSON.Send(c.connection, msg)

            case <-c.close:
                c.close <- true
                return
            }
        }
    }

    func (c *Client) listenToRead() {
        log.Println("Listening read from client")
        for {
            select {
            case <-c.close:
                c.close <- true
                return

            default:
                var msg Message
                err := websocket.JSON.Receive(c.connection, &msg)
                fmt.Printf("Received: %+v\n", msg)
                if err == io.EOF {
                    c.close <- true
                } else if err != nil {
                    // c.server.Err(err)
                } else {
                    broadcast(&msg)
                }
            }
        }
    }
    ...

Since we cannot have two blocking commands, one of them must be executed in a separate goroutine. In order to do that cleanly, we should create another method that will start listening for both reading and writing messages:

    ...
    func (c *Client) listen() {
        go c.listenToWrite()
        c.listenToRead()
    }
    ...

Last thing we need for our `Client` is a constructor so that we only need to pass an existing WS connection, while creating both channels inside:

    ...
    func NewClient(ws *websocket.Conn) Client {
        ch := make(chan *Message, 100)
        close := make(chan bool)

        return Client{ws, ch, close}
    }
    ...    

**Server:** Our server side needs to have a few methods and one important attribute - list of clients. Adding a new client will result in adding a new `Client` instance to the list. Also, once we get an established WS connection, we can greet it via `websocket.JSON.Send(..)` method using defined `Message` format:

    ...
    var clients []Client
    ...
    func addClientAndGreet(list []Client, client Client) []Client {
        clients = append(list, client)
        websocket.JSON.Send(client.connection, Message{"Server", "Welcome!"})
        return clients
    }
    ...    

Registration of a new connection is limited to creating a new `Client` item, passing it to `addClientAndGreet(..)` function and then executing `listen(..)` method in order to trigger those listening loops:

    ...
    func onWsConnect(ws *websocket.Conn) {
        defer ws.Close()
        client := NewClient(ws)
        clients = addClientAndGreet(clients, client)
        client.listen()
    }
    ...    

Since we don't want to reveal or implementation to the `main.go` file, we should assign our `onWsConnect` function to a `websocket.Handler`:

    var wsHandler = websocket.Handler(onWsConnect)

The last thing we need from our server is broadcasting messages. This is the simplest function we need to implement, as we just need to iterate over the list of `Client` objects and send it to each of them:

    ...
    func broadcast(msg *Message) {
        fmt.Printf("Broadcasting %+v\n", msg)
        for _, c := range clients {
            c.ch <- msg
        }
    }
    ...

**Back to the start:** To wrap things up, we need to call `broadcast` method once a user enters our broadcasting URL:

    // main.go
    func broadcastHandler(w http.ResponseWriter, r *http.Request) {
        msg := readMsgFromRequest(r)
        broadcast(&Message{"Server", msg})
        fmt.Fprintf(w, "Broadcasting %v", msg)
    }

Complete source code of this example is available [on Github](https://github.com/mycodesmells/golang-websockets).
