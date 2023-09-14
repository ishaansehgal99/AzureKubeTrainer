## failed to setup ring 

## errror

```shell
torch.distributed.DistBackendError: NCCL error in: /opt/pytorch/pytorch/torch/csrc/distributed/c10d/ProcessGroupNCCL.cpp:1197, internal error - please report this issue to the NCCL developers, NCCL version 2.18.3
ncclInternalError: Internal check failed.
Last error:
Error : ring 1 does not contain rank 0
    return _run_component(components, cfg_init)
  File "/usr/local/lib/python3.10/dist-packages/jsonargparse/_cli.py", line 181, in _run_component
    return component(**cfg)
  File "/host/home/azureuser/lit-llama/finetune/adapter_v2.py", line 103, in main
    model, optimizer = fabric.setup(model, optimizer)
  File "/usr/local/lib/python3.10/dist-packages/lightning/fabric/fabric.py", line 225, in setup
    CLI(main)
  File "/usr/local/lib/python3.10/dist-packages/jsonargparse/_cli.py", line 96, in CLI
    module, optimizers = self._strategy.setup_module_and_optimizers(  # type: ignore[assignment]
  File "/usr/local/lib/python3.10/dist-packages/lightning/fabric/strategies/deepspeed.py", line 326, in setup_module_and_optimizers
    return _run_component(components, cfg_init)
  File "/usr/local/lib/python3.10/dist-packages/jsonargparse/_cli.py", line 181, in _run_component
    self._deepspeed_engine, optimizer = self._initialize_engine(module, optimizers[0])
  File "/usr/local/lib/python3.10/dist-packages/lightning/fabric/strategies/deepspeed.py", line 583, in _initialize_engine
    return component(**cfg)
  File "/host/home/azureuser/lit-llama/finetune/adapter_v2.py", line 103, in main
        model, optimizer = fabric.setup(model, optimizer)deepspeed_engine, deepspeed_optimizer, _, _ = deepspeed.initialize(

  File "/usr/local/lib/python3.10/dist-packages/lightning/fabric/fabric.py", line 225, in setup
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/__init__.py", line 171, in initialize
    engine = DeepSpeedEngine(args=args,
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/runtime/engine.py", line 261, in __init__
    module, optimizers = self._strategy.setup_module_and_optimizers(  # type: ignore[assignment]
  File "/usr/local/lib/python3.10/dist-packages/lightning/fabric/strategies/deepspeed.py", line 326, in setup_module_and_optimizers
    self._configure_distributed_model(model)
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/runtime/engine.py", line 1118, in _configure_distributed_model    self._deepspeed_engine, optimizer = self._initialize_engine(module, optimizers[0])
  File "/usr/local/lib/python3.10/dist-packages/lightning/fabric/strategies/deepspeed.py", line 583, in _initialize_engine
        deepspeed_engine, deepspeed_optimizer, _, _ = deepspeed.initialize(self._broadcast_model()

  File "/usr/local/lib/python3.10/dist-packages/deepspeed/runtime/engine.py", line 1041, in _broadcast_model
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/__init__.py", line 171, in initialize
    engine = DeepSpeedEngine(args=args,
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/runtime/engine.py", line 261, in __init__
        dist.broadcast(p, groups._get_broadcast_src_rank(), group=self.seq_data_parallel_group)self._configure_distributed_model(model)

  File "/usr/local/lib/python3.10/dist-packages/deepspeed/comm/comm.py", line 117, in log_wrapper
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/runtime/engine.py", line 1118, in _configure_distributed_model    return func(*args, **kwargs)
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/comm/comm.py", line 224, in broadcast
    return cdb.broadcast(tensor=tensor, src=src, group=group, async_op=async_op)
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/comm/torch.py", line 192, in broadcast
    self._broadcast_model()
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/runtime/engine.py", line 1041, in _broadcast_model
    return torch.distributed.broadcast(tensor=tensor, src=src, group=group, async_op=async_op)
  File "/usr/local/lib/python3.10/dist-packages/torch/distributed/c10d_logger.py", line 47, in wrapper
    return func(*args, **kwargs)
  File "/usr/local/lib/python3.10/dist-packages/torch/distributed/distributed_c10d.py", line 1884, in broadcast
    dist.broadcast(p, groups._get_broadcast_src_rank(), group=self.seq_data_parallel_group)
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/comm/comm.py", line 117, in log_wrapper
    return func(*args, **kwargs)
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/comm/comm.py", line 224, in broadcast
    return cdb.broadcast(tensor=tensor, src=src, group=group, async_op=async_op)
  File "/usr/local/lib/python3.10/dist-packages/deepspeed/comm/torch.py", line 192, in broadcast
    work = group.broadcast([tensor], opts)
```

## solution
run pytorch job with following envs
export NCCL_DEBUG=1
export NCCL_P2P_DISABLE=1
