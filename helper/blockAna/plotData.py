import matplotlib.pyplot as plt

labels = [
    "L", "B", "DM", "DC",
    "L+B", "L+DM", "L+DC",
    "B+DM", "B+DC", "DM+DC",
    "L+B+DM", "L+B+DC",
    "L+DM+DC", "B+DM+DC",
    "L+B+DM+DC"
]


def plot_results(data: list[int]):
    print("Plot Data")
    plt.figure(figsize=(12, 6))

    x = range(len(data))
    plt.bar(x, data)

    plt.ylabel("Number")
    plt.xticks(x, labels, rotation=22.5)

    plt.tight_layout()
    plt.savefig("numberBugs.png")
    plt.close()

def plot_len_cycle(data: dict[int, int]):
    print("Plot Len Cycle")

    x = range(2, max(data.keys()) + 1)
    data_list = []
    for i in x:
        if i in data.keys():
            data_list.append(data[i])
        else:
            data_list.append(0)


    plt.figure(figsize=(12, 6))

    plt.bar(x, data_list)

    plt.xlabel("Length")
    plt.ylabel("Number")
    # plt.xticks(x, labels, rotation=22.5)

    plt.tight_layout()
    plt.savefig("lenCycle.png")
    plt.close()

