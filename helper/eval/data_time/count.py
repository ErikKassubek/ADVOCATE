import sys
import csv


def sum_second_column(file_path):
    total = 0
    try:
        with open(file_path, mode='r') as file:
            reader = csv.reader(file)
            for row in reader:
                if len(row) > 1:  # Ensure the row has at least two elements
                    try:
                        # Convert to float and add to total
                        total += float(row[8])
                    except ValueError:
                        print(f"Skipping invalid value: {row[8]}")
    except FileNotFoundError:
        print(f"File not found: {file_path}")
    return total


# Example usage
file_path = sys.argv[1]  # Replace with the path to your CSV file
result = sum_second_column(file_path)
print(f"The sum of the second column is: {result}")
