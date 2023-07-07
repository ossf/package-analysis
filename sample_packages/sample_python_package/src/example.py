import http.client
import json

# Sends an HTTPS post request and prints out the response.
def sendHTTPSPostRequest():
  host = "www.httpbin.org"
  conn = http.client.HTTPSConnection(host)
  data = {'text': 'Sending data through HTTPS'}
  json_data = json.dumps(data)
  conn.request("POST", "/post", json_data, headers={"Host": host})
  response = conn.getresponse()
  print(response.read().decode())

def main():
  sendHTTPSPostRequest()


if __name__ == "__main__":
  main()
