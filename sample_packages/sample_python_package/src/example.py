import http.client
import json

# Sends an HTTPS post request and prints out the response.
def send_https_post_request(location: str) -> None:
  host = "www.httpbin.org"
  conn = http.client.HTTPSConnection(host)
  data = {'text': 'Sending data through HTTPS from: ' + location}
  json_data = json.dumps(data)
  conn.request("POST", "/post", json_data, headers={"Host": host})
  response = conn.getresponse()
  print(response.read().decode())

def main():
  send_https_post_request("main function")


if __name__ == "__main__":
  main()
