import socket

HOST = 'localhost'  # The server's hostname or IP address
PORT = 8080        # The port used by the server

while True:
        try:
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                        s.connect((HOST, PORT))
                        while True:
                                # Get user input
                                user_input = input("Prompt(name^index): ")
                                # Send the user input to the server
                                s.sendall(user_input.encode())
        except:
                continue
