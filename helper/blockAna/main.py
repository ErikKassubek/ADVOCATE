from analyseData import analyze_data
from readData import read_data
from plotData import plot_data

numTest = {
    "argo-cd": 1918,
    "bleve": 582,
    "bosun": 167,
    "caddy": 267,
    "dns": 297,
    "flannel": 39,
    "frp": 32,
    "gin": 528,
    "go-ethereum": 1648,
    "gofiber": 522,
    "gorums": 39,
    "grpc": 226,
    "kubernetes": 6403,
    "nsq": 160,
    "ollama": 292,
    "prometheus": 1077,
    "terraform": 3216,
    "zinx": 38,
    "sum": 17451
}

def main():
    path = "/home/erik/Uni/Programs/"
    data = read_data(path)
    res_total, res_per_prog = analyze_data(data)
    plot_data(res_total)
    
    printResults(res_per_prog)
    print(res_total)


def printResults(res: dict[str, list[int]]):
    k = list(res.keys())
    k.sort()
    for prog in k:
        print("| " + prog + " | " + str(numTest[prog]) + " | " + " | ".join(str(num) for num in res[prog]) + " |")


if __name__ == "__main__":
    main()