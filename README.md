
# Code of paper "A Scheduling Method for Tasks and Services in IIoT Multi-Cloud Environments"

Link to paper: https://ieeexplore.ieee.org/document/10257193

---

## Implementation of Algorithms
* **First Fit**: "algorithms/first_fit.go"
* **Random Fit**: "algorithms/random_fit.go"
* **NSGA-II**: "algorithms/nsga2.go"
* **HAGA**: "algorithms/haga.go"
* **MCASGA**: "algorithms/genetic.go"

---

## Experiments

### The paper's Subection IV. EVALUATION B. Weaken the Influence of Random Factors
* **Experiment code**: "experiments/validateaverage/main.go"
* **Experiment data**: "experiments-data/validate average"

### The paper's Subection IV. EVALUATION C. MCASGA Operators
* **Experiment code**: "experiments/optimizecpmp/main.go"
* **Experiment data**: "experiments-data/Fitness values of different operators.txt"

### The paper's Subection IV. EVALUATION D. Experiments and Metrics, E. Results
* **Experiment code**: "experiments/continuousexperiment/main.go"
* **Experiment data**:
    + _0% Tasks and 100% Services_: "experiments-data/csv-and-matlab/0"
    + _25% Tasks and 75% Services_: "experiments-data/csv-and-matlab/25"
    + _50% Tasks and 50% Services_: "experiments-data/csv-and-matlab/50"
    + _75% Tasks and 25% Services_: "experiments-data/csv-and-matlab/75"
    + _100% Tasks and 0% Services_: "experiments-data/csv-and-matlab/100"
