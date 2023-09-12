# Run this file on kubernetes with 
# python test.py run --with kubernetes --parallel=10
from metaflow import FlowSpec, Parameter, step
import time

class MLDemoFlow(FlowSpec): 
    parallel = Parameter("parallel", help = "how many processes to run training", default=10)
    
    @step
    def start(self):
        self.models = list(range(self.parallel))
        # Would spawn "parallel" # of pods
        self.next(self.train, foreach='models')
    @step
    def train(self):
        print("Training process number %d" % self.index)
        # Perform training here
        time.sleep(30)
        self.next(self.join)
    
    @step
    def join(self, inputs): 
        print("%d takss completed successfully!" % len(list(inputs)))
        self.next(self.end)
    
    @step
    def end(self):
        print("Finished training!")
        