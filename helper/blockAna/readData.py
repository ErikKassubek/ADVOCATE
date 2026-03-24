import os
from blockInfo import BlockInfo

UNIQUE = True


res_total = []

def read_data(path) -> dict[str, dict[str, list[BlockInfo]]]:
    res = {}

    for entry in os.listdir(path):
        full_path = os.path.join(path, entry)
        if os.path.isdir(full_path):
            data = read_data_prog(full_path, entry)
            res[entry] = data

    return res


def read_data_prog(path: str, name: str) -> dict[str, list[BlockInfo]]:
    print("Read Data: ", name)
    res_total = []
    result_path = os.path.join(path, "advocateResult")

    res = {}

    if not os.path.exists(result_path):
        return res

    for test in os.listdir(result_path):
        test_path = os.path.join(result_path, test)
        if not os.path.isdir(test_path):
            continue

        key = f"{name}->{test}"
        res[key] = read_data_test(test_path)

    return res


def read_data_test(path) -> list[BlockInfo]:
    path_result = os.path.join(path, "bugs")

    if not os.path.exists(path_result):
        return []

    res = {}

    for bug in os.listdir(path_result):
        res_path = os.path.join(path_result, bug)
        info = BlockInfo()
        run = bug.split("_")[1]
        if run in res.keys():
            info = res[run]
        type_id = ""
        with open(res_path, "r") as f:
            for line in f:
                line = line.strip()

                if not line:
                    continue

                if line.startswith("# "):
                    type_id = line.split(" ")[2]

                    if len(type_id) != 3:
                        raise ValueError("Invalid type id: ", type_id, " ", path)

                    if type_id.startswith("L") and type_id != "L00":
                        info.set = True
                        info.leak = True
                    elif type_id == "A07":
                        info.set = True
                        info.block = True
                    elif type_id == "A08":
                        info.set = True
                        info.deadlock_mutex = True
                    elif type_id == "A10":
                        info.set = True
                        info.deadlock_mixed = True
                elif line.startswith("->"):
                    posStr = line.removeprefix("-> ")
                    if not posStr.startswith("/home/erik/Uni/Advocate/goPatch"):
                        info.pos[posStr] = info.deadlock_mutex or info.deadlock_mixed


        if info.set and len(info.pos.keys()) > 0 and (not UNIQUE or not contains(info)):
            print(run, " -> ", type_id, " -> ", info.pos.keys())
            res[run] = info
            res_total.append(info)

    return res.values()



def contains(info: BlockInfo):
    for res in res_total:
        if info.isEqual(res):
            return True
        
    return False

