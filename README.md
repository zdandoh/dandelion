```
Functions:
func_name = f(){

}

Parens can be omitted with no arguments. Three arguments are automatically passed:
e - the element that the function is operating on
i - the element number that the function is operating on
a - the array that the function is operating on

Function modifiers:

f{} - Normal function
fi{} - Filter
fm{} - Map, any modifications to e are kept

Pipelines:
array -> function -> function
Double pipeline notation:
array ->> function ->> p
unrolls a multi-dimensional array and addresses each element with the function

Data Types:
bytes
function
int64
float64

Built-ins:

Functions:
rf("file_name") - reads all the files from a directory or reads the contents of a file
df - deletes a file
p - prints an array to stdout
len - turns an array into a single element numerical array

Variables:
pwd - current working directory

Toy Programs:
files = rf()
my_func = fm{
    e + 1
}
[1, 2, 3, 4] -> my_func -> p
rf() -> p
rf(pwd) -> p
rf() -> fi{ "b" in e } -> df
["butt", "nut", "futt"] -> fi{ "b" in e } -> p
[0..100]-> f{
    e % 5 == 0: p()
    e % 2 == 0: pfizz
}

create variable, call function, assign to variable
function call, pipe, function call
function call, pipe, function call
function call, pipe, function definition & call, function call
literal definition, pipe, function definition & call, pipe, function call
```
