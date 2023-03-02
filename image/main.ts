const SERVICE = Deno.env.get("K_SERVICE");
const REVISION = Deno.env.get("K_REVISION");

async function serveHttp(conn: Deno.Conn) {
  const httpConn = Deno.serveHttp(conn);

  for await (const event of httpConn) {
    const request = event.request;
    console.log(`Received request for: ${request.method} ${request.url}`);

    const body = {
        service: SERVICE,
        revision: REVISION,
    };

    event.respondWith(
      Response.json(body, {
        status: 200,
      }),
    );
  }
}

const port = 8080;
const server = Deno.listen({ port });
console.log(`Running on :${port}`);

for await (const conn of server) {
  serveHttp(conn);
}
