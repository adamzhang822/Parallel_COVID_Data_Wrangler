# Parallel_COVID_Data_Wrangler

This project implements a parallelized COVID data processor to simulate the process of parallel pre-processing of datasets that may contain duplicate or invalid records.
The program takes as input a zipcode, a month, and a year specified by the user, parses through ≥ 500 csv files with potential duplications and invalidate entries to gather and generate the total number of COVID cases, tests, and deaths for the specified zipcode over the specified month and year.

For more details, refer to the Writeup_Final.pdf in the repository.

# Data

The datasets used come from City of Chicago Data Portal's COVID-19 Cases, Tests, and Deaths by ZIP Code dataset. 
The original unmodified data from the source with unique entries are stored in covid_sample.csv .
For the program, we will use modified 500 data files named covid_NUM.csv , where O<=NUM<=500 , each consisting of about ~37,000 random lines sampled with replacement from the source file. The 500 files are generated in such a way that it is guaranteed that
together they contain all entries from the source file. 
In some test cases, we will use more than 500 files. In those cases, we will simply recycle one of the 500 files for each file number greater than 500.

# Parallel Implementations:
The program has a sequential implementation and 3 parallel implementations, which can be toggeled by specifying them in the usage statement below.
```
const usage =
    "Usage: go run proj3/covid mode size threads zipcode month year\n" +
    " mode = either 'static' or 'stealing' or 'bsp'\n" +
    " size = 500 or 1000 or 3000, the number of files to be processed\n" +
    " threads = the number of threads (i.e., goroutines to spawn).
                If bsp, must be > 2 \n" +
    " to run sequential mode, specify thread = 0 when the mode is either static or stealing\n" +
    " zipcode = a possible Chicago zipcode\n" +
    " month = the month to display for that zipcode, must be between 1-12 \n" +
    " year  = the year to display for that zipcode, must be 2020 or 2021 \n"
```

For details about each parallel implementations, please refer to the system writeup in Writeup_Final.pdf

# Running the program:
The program can be ran following the usage statements provided in the section above. In the proj3/covid directory, run the following commands to produce the results below:

```
$: go run proj3/covid static 500 0 60603 5 2020
2,48,0
$: go run proj3/covid bsp 3000 4 60640 2 2021
182,9961,5
$; go run proj3/covid static 1000 3 89149 2 2020
0,0,0
$; go run proj3/covid stealing 500 3 89149 2 2020
0,0,0
```

Since data is read in using path in the file, you must be in the folder to run the program. Otherwise, the program will fail to get the data.
