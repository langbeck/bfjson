# What is `bfjson`?
`bfjson` was a side project that I've created to showcase how JSON parsing could be improved in the company I did worked.
It's a module-aware CLI tool that generates custom JSON decoders with many extra features (such as unsafe strings) that can improve the performance even further.

# What's the state of `bfjson`?
This is only the proof-of-concept tool that I'd created before it was incorported in the product.
It's rudimental but over time I'll add features that doesn't conflict with the proprietary solution.
For now I'm using this to showcase how AST parsing could be used for code generation at build time can be better than using reflection-based solutions during runtime.

# Usage
We could use `bfjson` to generate custom decoders using both supported engines for the package `github.com/bsm/openrtb` (which is a classic package for Real-Time Bidding applications) using the code bellow:
```bash
mkdir -p codecs/fastjson && bfjson -pkg github.com/bsm/openrtb -engine fastjson -pkgname fastjson -write codecs/fastjson/codec.go
mkdir -p codecs/custom && bfjson -pkg github.com/bsm/openrtb -engine custom -pkgname custom -write codecs/custom/codec.go
```

# External references
This tool have support for a custom engine (based on Dave Cheney's `github.com/pkg/json`) and also for `github.com/valyala/fastjson`.

# Known issues
- Doesn't support unescaping quoted strings
