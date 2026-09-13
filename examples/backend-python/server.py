from http.server import BaseHTTPRequestHandler, HTTPServer


class EventHandler(BaseHTTPRequestHandler):
    def do_POST(self) -> None:
        content_length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(content_length)

        print(body.decode("utf-8"))

        self.send_response(204)
        self.end_headers()


if __name__ == "__main__":
    server = HTTPServer(("127.0.0.1", 9100), EventHandler)
    print("Example Python backend listening on http://127.0.0.1:9100/events")
    server.serve_forever()
