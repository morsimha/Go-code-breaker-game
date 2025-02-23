package server;

import java.util.Random;

public class GameLogic {
    private static final int SECRET_CODE = 1111;
    private final Random random = new Random();

    public static Result validateGuess(String input) {
        try {
            int guess = Integer.parseInt(input); // Convert string to integer
            return new Result(guess, null); // Return the guess and no error
        } catch (NumberFormatException e) {
            return new Result(-1, "Invalid input: Not a valid integer"); // Return error message
        }
    }

    public int generateSecretCode() {
        return SECRET_CODE;
    }

    // GenerateTimestampPrefix generates a textual prefix containing the current time
    public static String generateTimestampPrefix() {
        long timestamp = System.currentTimeMillis() / 1000; // Convert to seconds
        String prefix = "TIME: " + timestamp;
        new Thread(() -> {
            String message = String.format("this is my prefix: %s", prefix);
        }).start();
        return prefix;
    }
}

