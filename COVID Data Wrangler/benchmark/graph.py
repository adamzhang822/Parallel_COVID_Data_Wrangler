import subprocess 
import matplotlib.pyplot as plt
import timeit

# preps 
print("starting")
threads = [2, 4, 6, 8, 12]
sizeList = ["500", "1000", "3000"]

def execute_scripts(argString):
    args = argString.split(" ")
    print("executing {} \n".format(argString))
    subprocess.call(['go', 'run', 'proj3/covid', args[0], args[1], args[2], "60603", "5", "2020"])

def execute_serial(size):
    print("executing sequential for size {} \n".format(size))
    subprocess.call(['go', 'run', 'proj3/covid', "static", size, "0", "60603", "5", "2020"])

serialSpeeds = {}
for size in sizeList:
    executable_serial = 'execute_serial("{input}")'.format(input = size)
    serialTimer = timeit.Timer(executable_serial, setup="from __main__ import execute_serial")
    serialSpeeds[size] = serialTimer.timeit(5) / 5
    print(str(serialSpeeds[size]) + "\n")

def generatePlot(mode):
    print("in {} mode \n".format(mode))
    speedups = {}
    for size in sizeList:
        sizeSpeedups = []
        for thread in threads:
            workers = 0
            if mode == "bsp":
                workers = thread + 1
            else:
                workers = thread
            argString = " ".join([mode, size, str(workers)])
            print(argString + "\n")
            executable = 'execute_scripts("{input}")'.format(input = argString)
            t = timeit.Timer(executable, setup = "from __main__ import execute_scripts")
            speedup = serialSpeeds[size] / (t.timeit(5) / 5)
            print(str(speedup) + "\n")
            sizeSpeedups.append(speedup)
        speedups[size] = sizeSpeedups
        
    # plotting graphs
    for size in sizeList:
        plt.plot(threads, speedups[size], label = size)
    plt.xlabel('Number of Threads')
    plt.ylabel('Speedup')
    plt.title("Covid Speedups Graph ({})".format(mode))
    plt.legend()
    plt.savefig('speedup-{}.png'.format(mode))
    plt.clf()
    return 0

generatePlot("static")
generatePlot("stealing")
generatePlot("bsp")



