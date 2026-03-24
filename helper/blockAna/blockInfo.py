class BlockInfo:
    def __init__(self):
        self.set = False
        self.leak = False
        self.block = False
        self.deadlock_mutex = False
        self.deadlock_mixed = False
        self.pos = {}

    def as_number(self):
        fields = [
            self.leak,
            self.block,
            self.deadlock_mutex,
            self.deadlock_mixed,
        ]

        result = []
        n = len(fields)

        def comb(start, k, current):
            if len(current) == k:
                all_true = all(fields[i] for i in current)
                result.append(1 if all_true else 0)
                return

            for i in range(start, n):
                comb(i + 1, k, current + [i])

        for k in range(1, n + 1):
            comb(0, k, [])

        return result

    def isEqual(self, other) -> bool:
        if not self.set or not other.set:
            return False
        
        if self.leak != other.leak:
            return False 
        
        if self.block != other.block:
            return False 
        
        if self.deadlock_mutex != other.deadlock_mutex:
            return False 
        
        if self.leak != other.deadlock_mixed:
            return False 
        
        if posString(self.pos) != posString(other.pos):
            return False 
        
        return True
    

def posString(pos: dict):
    resList = []
    for key in pos.keys():
        resList.append(key)
    
    resList.sort()
    resStr = "-".join(resList)

    return resStr