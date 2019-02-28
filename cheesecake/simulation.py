import abc
from .models import *
import pandas as pd
import numpy as np

class EventSimulator(abc.ABC):
    def __init__(self, event, predictor):
        self.event = event
        self.predictor = predictor

    def matches(self):
        return

    # TODO
    #def awards(self):
    #    return

    def rankings(self):
        return

class PreEventSimulator(EventSimulator):
    def matches(self):
        """
        Matches have not been announced yet, so we cannot
        predict them quite yet.
        """
        return []

    def rankings(self):
        """
        Matches have not been announced, create simulations.
        """
        teams = [x.key for x in self.event.teams]
        np.random.seed(0)
        sample = np.random.choice(teams, size=(10000, 6))
        predictions = [self.predictor.predict_keys(x) for x in sample]
        reds = sample[:,0:3].flatten()
        blues = sample[:,3:6].flatten()
        pred_repeat = np.repeat(predictions, 3)
        df = pd.DataFrame({
            "teams": np.concatenate((reds, blues), axis=None),
            "predictions": np.concatenate((pred_repeat, np.subtract(1.0, pred_repeat)))
        })
        dic = df.groupby("teams").mean().sort_values(by='predictions', ascending=False)
        ## TODO: Surely there is a better way to do this.
        values = []
        for key, val in dic["predictions"].iteritems():
            values.append({
                'key': key,
                'mean': val
            })
        return values

class QualificationEventSimulator(EventSimulator):
    def matches(self):
        return [x.predictions for x in self.event.matches]


    def rankings(self):
        return []
