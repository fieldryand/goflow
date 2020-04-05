# DAG validation

Let's step through the algorithm to check if a dag is acyclic.

## Acyclic case

```
d.graph = [A: [B, C], B: [D], C: [D], D: []]
```

Turn the graph sideways--degree is the depth in the hierarchy.

```
degree = {A: 0, B: 1, C: 1, D: 2}
```

The final for loop:

```
deq = (A)

1. l = [A]
   degree = {A: 0, B: 0, C: 0, D: 2}
   deq = (C, B)

2. l = [A, B]
   degree = {A: 0, B: 0, C: 0, D: 1}
   deq = (C)

3. l = [A, B, C]
   degree = {A: 0, B: 0, C: 0, D: 0}
   deq = (D)

4. l = [A, B, C, D]
   degree = {A: 0, B: 0, C: 0, D: 0}
   deq = ()

5. break

len(l) == len(d.graph) --> true, so it's acyclic
```

## Cyclic case

```
d.graph = [A: [B, C], B: [D], C: [D], D: [C]]

degree = {A: 0, B: 1, C: 2, D: 2}
```

The final for loop:

```
deq = (A)

1. l = [A]
   degree = {A: 0, B: 0, C: 1, D: 2}
   deq = (C, B)

2. l = [A, B]
   degree = {A: 0, B: 0, C: 1, D: 1}
   deq = (C)

3. l = [A, B, C]
   degree = {A: 0, B: 0, C: 1, D: 0}
   deq = (D)

4. l = [A, B, C, D]
   degree = {A: 0, B: 0, C: 0, D: 0}
   deq = (C)

5. l = [A, B, C, D, C]
   degree = {A: 0, B: 0, C: 0, D: 0}
   deq = ()

6. break

len(l) == len(d.graph) --> false, so it's cyclic
```
