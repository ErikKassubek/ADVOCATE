import matplotlib.pyplot as plt

labels = [
    "L", "B", "DM", "DC",
    "L+B", "L+DM", "L+DC",
    "B+DM", "B+DC", "DM+DC",
    "L+B+DM", "L+B+DC",
    "L+DM+DC", "B+DM+DC",
    "L+B+DM+DC"
]


def plot_data(data):
    print("Plot Data")
    plt.figure(figsize=(12, 6))

    x = range(len(data))
    plt.bar(x, data)

    plt.ylabel("Number")
    plt.xticks(x, labels, rotation=22.5)

    plt.tight_layout()
    plt.savefig("barchart.png")
    plt.close()