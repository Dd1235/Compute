import requests

base_url = "http://localhost:8090"
response = requests.get(base_url + "/ping")


print("======================Try using Requests====================")
print(f"Got response {response}")
print(response.headers)
print(response.history)
print(response.connection)
print(response.text)


print("===============================================================\n\n")

import http.client

print("======================Try using HTTP Client====================")


# http.client.HTTPConnection expects a hostname (no scheme) and an optional port.
# Passing "http://localhost" makes the socket library try to resolve the scheme as part
# of the host and causes socket.gaierror: [Errno 8] nodename nor servname provided, or not known.
conn = http.client.HTTPConnection("localhost", 8090)
# use the same path used above ("/ping") so both requests match the same endpoint
conn.request("GET", "/ping")
response = conn.getresponse()
print(f"HTTP Status: {response.status}, Reason: {response.reason}")

print("===============================================================\n\n")
