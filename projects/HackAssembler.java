import java.io.File;
import java.util.HashMap;
import java.util.Map;
import java.util.Scanner;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class HackAssembler {

    static final boolean DEBUG = false;

    static Map<String, Integer> LABELS = new HashMap<>(); // label -> index in ROM (code)
    static Map<String, Integer> VARS = new HashMap<>(); // var -> index in RAM (data)

    static Map<String, Integer> CONST = new HashMap<>() {{
        put("R0", 0);
        put("R1", 1);
        put("R2", 2);
        put("R3", 3);
        put("R4", 4);
        put("R5", 5);
        put("R6", 6);
        put("R7", 7);
        put("R8", 8);
        put("R9", 9);
        put("R10", 10);
        put("R11", 11);
        put("R12", 12);
        put("R13", 13);
        put("R14", 14);
        put("R15", 15);
        put("SP", 0);
        put("LCL", 1);
        put("ARG", 2);
        put("THIS", 3);
        put("THAT", 4);
        put("SCREEN", 16384);
        put("KBD", 24576);
    }};

    static Map<String, String> COMP = new HashMap<>() {{
        put("0", "0101010");
        put("1", "0111111");
        put("-1", "0111010");
        put("D", "0001100");
        put("A", "0110000");
        put("M", "1110000");
        put("!D", "0001101");
        put("!A", "0110001");
        put("!M", "1110001");
        put("-D", "0110001");
        put("-A", "0110011");
        put("D+1","0011111");
        put("A+1", "0110111");
        put("M+1", "1110111");
        put("D-1", "0001110");
        put("A-1", "0110010");
        put("M-1", "1110010");
        put("D+A", "0000010");
        put("D+M", "1000010");
        put("D-A", "0010011");
        put("D-M", "1010011");
        put("A-D", "0000111");
        put("M-D", "1000111");
        put("D&A", "0000000");
        put("D&M", "1000000");
        put("D|A", "0010101");
        put("D|M", "1010101");
    }};

    static Map<String, String> DEST = new HashMap<>() {{
        put("", "000");
        put("M", "001");
        put("D", "010");
        put("MD", "011");
        put("A", "100");
        put("AM", "101");
        put("AD", "110");
        put("AMD", "111");
    }};

    static Map<String, String> JMP = new HashMap<>() {{
        put("", "000");
        put("JGT", "001");
        put("JEQ", "010");
        put("JGE", "011");
        put("JLT", "100");
        put("JNE", "101");
        put("JLE", "110");
        put("JMP", "111");
    }};


    static int VAR_INDEX = 16;


    public static void main(String[] args) throws Exception {
        if (args.length != 1) {
            System.out.println("need path to asm file");
            System.exit(1);
        }

        Scanner scanner;

        scanner = new Scanner(new File(args[0]));

        // first pass, fill in the labels

        int line;

        line = 0;
        int runningL = 0;
        while (scanner.hasNext()) {
            String s = scanner.nextLine().trim();
            line++;
            if (s.isBlank() || s.startsWith("/")) {
                continue;
            }
            if (!s.startsWith("(")) {
                runningL++;
                continue;
            }
            String labelName = labelName(s);
            LABELS.put(labelName, runningL );
        }


        // second pass
        line = 0;
        scanner = new Scanner(new File(args[0]));

        while (scanner.hasNext()) {
            String s = scanner.nextLine();
            line++;
            if (s.startsWith("//")) {
                continue;
            }
            s = s.trim();
            if (s.length() < 1) {
                continue;
            }
            String cmd;
            switch (s.charAt(0)) {
                case '@':
                    cmd = instrA(s.substring(1));
                    break;
                case '(':
                    continue;
                default:
                    cmd = instrC(s);
            }
            if (DEBUG) {
                System.out.printf("line %d: ", line);
            }
            assert cmd.length() == 16;
            System.out.println(cmd);
        }

    }

    private static String labelName(String s) {
        Matcher matcher = Pattern.compile("\\((.+)\\)").matcher(s);
        if (!matcher.find()) {
            System.err.println("Error parsing " + s);
            System.exit(1);
        }
        return matcher.group(1);
    }

    private static String instrA(String in) {
        if (CONST.containsKey(in)) {
            return bool16(CONST.get(in));
        }
        if (LABELS.containsKey(in)) {
            return bool16(LABELS.get(in));
        }
        if (Character.isDigit(in.charAt(0))) {
            return bool16(Integer.valueOf(in));
        }
        if (!VARS.containsKey(in)) {
            VARS.put(in, VAR_INDEX);
            VAR_INDEX++;
        }
        return bool16(VARS.get(in));
    }

    private static String bool16(Integer integer) {
        String s = "000000000000000" + Integer.toBinaryString(integer);
        return s.substring(s.length() - 16);
    }

    private static String label(String in) {
        String label = labelName(in);
        return String.valueOf(LABELS.get(label));
    }

    private static String instrC(String in) {
        Matcher matcher;
        matcher = Pattern.compile("(\\w+)=(\\S+)").matcher(in); // [0] -> dest, [1] -> comp
        if (matcher.find()) {
            String hex = "111" + comp(matcher.group(2)) + dest(matcher.group(1)) + jump("");
            return hex;
        }
        matcher = Pattern.compile("(\\w+);(\\w+)").matcher(in); // [0] -> comp, [1] -> jump
        if (matcher.find()) {
            String hex = "111" + comp(matcher.group(1)) + dest("")  + jump(matcher.group(2));
            return hex;
        }

        return null;

    }

    private static String dest(String s) {
        return DEST.get(s);
    }

    private static String comp(String s) {
        return COMP.get(s);
    }

    private static String jump(String s) {
        return JMP.get(s);
    }


}
