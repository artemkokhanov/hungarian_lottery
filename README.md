Checking correctness using the README.md file:
Running the program: go run main.go test_input.txt

                                       2 3 4 5
                                      ---------
Query: 1 2 3 4 5    → Expected output: 0 1 1 2
Query: 1 9 10 11 12 → Expected output: 0 0 0 1
Query: 2 9 13 14 15 → Expected output: 0 0 0 1
Query: 10 20 6 8 22 → Expected output: 0 0 0 1
