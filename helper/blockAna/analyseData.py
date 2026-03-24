from blockInfo import BlockInfo, posString

def add_slices(slice1, slice2):
    if len(slice1) != len(slice2):
        raise ValueError("Slices have different lengths")

    for i in range(len(slice1)):
        slice1[i] += slice2[i]


def analyze_data(data: dict[str, dict[str, list[BlockInfo]]]) -> tuple[list[int], dict[str, list[int]]]:
    res_per_prog = {}
    res_total = [0] * 15

    for prog, data_per_prog in data.items():
        print("Analyze: ", prog)
        res_per_prog[prog] = [0] * 15

        for data_per_test in data_per_prog.values():
            for data_per_run in data_per_test:
                numbers = data_per_run.as_number()
                add_slices(res_per_prog[prog], numbers)
                add_slices(res_total, numbers)

    return res_total, res_per_prog
