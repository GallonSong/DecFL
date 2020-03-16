import json
import sys
import click
import numpy as np
import os
from os import walk


@click.command()
@click.argument('w', type=click.STRING)
@click.argument('o', type=click.STRING)
def run_standalone(w, o):

    updates = []

    for file in absoluteFilePaths(w):
        with open(file, "r") as weights_file:
            weights = weights_file.read()
            weights = json.loads(weights)
            weights = [np.array(w) for w in weights]

        if weights is None:
            print("Weights could not be loaded, abort.", file=sys.stderr)
            sys.exit(1)
        else:
            print("Weights loaded from", file)

        if not checkWeightsFormat(weights):
            print("Weights could not be parsed correctly.", file=sys.stderr)
            sys.exit(1)

        updates.append(weights)

    new_weights = aggregateUpdates(updates)

    new_weights = [w.tolist() for w in new_weights]
    new_weights = json.dumps(new_weights)

    with open(o, "w") as output_file:
        output_file.write(new_weights)
        print("Saved updated weights to", o)

    #sys.exit(1)


def aggregateUpdates(updates):
    weights = [0*w for w in updates[0]]

    for update in updates:
        weights = [w1 + w2 for w1, w2 in zip(weights, update)]

    weights = [w/len(updates) for w in weights]
    weights = [np.array(w) for w in weights]

    return weights


def checkWeightsFormat(weights):
    if not isinstance(weights, list):
        return False
    if not len(weights) > 0:
        return False
    if not isinstance(weights[0], np.ndarray):
        return False
    if not len(weights[0]) > 0:
        return False
    return True


def absoluteFilePaths(directory):
    for dirpath, _, filenames in os.walk(directory):
        for f in filenames:
            yield os.path.abspath(os.path.join(dirpath, f))


if __name__ == '__main__':
    run_standalone()