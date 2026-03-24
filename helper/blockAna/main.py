from analyseData import analyze_data
from readData import read_data
from plotData import plot_data


def main():
    path = "/home/erik/Uni/Programs/"
    data = read_data(path)
    res_total, res_per_prog = analyze_data(data)
    printResults(res_per_prog)
    print(res_total)
    


    plot_data(res_total)

def printResults(res: dict[str, list[int]]):
    k = list(res.keys())
    k.sort()
    for prog in k:
        print("| " + prog + " | " + " | ".join(str(num) for num in res[prog]) + " |")


if __name__ == "__main__":
    main()