package server;

import server.GameLogic;
import java.io.*;
import java.net.*;

class ClientHandler extends Thread {
    private final Socket clientSocket;
    private final GameLogic gameLogic;

    public ClientHandler(Socket socket) {
        this.clientSocket = socket;
        this.gameLogic = new GameLogic();
    }

    public void run() {
        try (BufferedReader in = new BufferedReader(new InputStreamReader(clientSocket.getInputStream()));
             PrintWriter out = new PrintWriter(clientSocket.getOutputStream(), true)) {

            out.println("Welcome to the Code Breaker Game! Enter a code between 1000 and 9999.");
            String inputLine;
            int secretCode = gameLogic.generateSecretCode();

            while ((inputLine = in.readLine()) != null) {
                try {
                    out.println("Enter your guess (secret code) or 'exit' to quit: ");
                    int guess = gameLogic.validateGuess(inputLine);
                    String prefix = "";
//                     String prefix = gameLogic.generateTimestampPrefix();

                    if (secretCode == guess) {
                        out.println(prefix + " Congratulations! You guessed correctly!");
                        break;
                    } else {
                        out.println(prefix + " Try again!");
                    }
                } catch (IllegalArgumentException e) {
                    out.println(e.getMessage());
                }
            }
        } catch (IOException e) {
            e.printStackTrace();
        } finally {
            try {
                clientSocket.close();
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }
}
