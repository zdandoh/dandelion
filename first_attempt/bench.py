import time

t = time.time()
i = 0
iter = 10000000
while i < 10000000:
    i += 1
print((time.time() - t) / 10000000 * 1e9)
