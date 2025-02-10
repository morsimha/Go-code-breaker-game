package client;

import java.io.*;
import java.net.*;

public class GameClient {
    public static void main(String[] args) {
        String host = "localhost";
        int port = 8080;
        try (Socket socket = new Socket(host, port);
             BufferedReader in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
             PrintWriter out = new PrintWriter(socket.getOutputStream(), true);
             BufferedReader userInput = new BufferedReader(new InputStreamReader(System.in))) {

            System.out.println("Connected to Game Server");
            System.out.println(in.readLine());
            String userInputLine;
            while ((userInputLine = userInput.readLine()) != null) {
                out.println(userInputLine);
                System.out.println("Server: " + in.readLine());
                if ("exit".equalsIgnoreCase(userInputLine)) {
                    break;
                }
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
