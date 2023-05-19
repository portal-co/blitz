def reader_ctx(ctx):
    def action(deps,name,cmd,path):
        def result(cfg):
            e = {}
            for k in deps:
                e[k] = deps[k].run_with_cfg(cfg)
            return ctx.action(e,name,cmd,path)
        return struct(run_with_cfg = result)
    def join(x):
        def result(cfg):
            return ctx.join(x.run_with_cfg(cfg))
        return struct(run_with_cfg = result)
    def bind(tgt,id,run):
        def result(cfg):
            id2 = id + "_" + json.encode(cfg)
            def go(key):
                return run(key).run_with_cfg(cfg)
            return ctx.bind(tgt.run_with_cfg(cfg),id2,go)
        return struct(run_with_cfg = result)
    return struct(action = action, join = join, bind = bind)