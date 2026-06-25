# Worker Nodes & The Activity Registry

The Worker is "The Brawn" of Chronos. It is completely stateless and designed to be horizontally scaled (e.g., running 50 identical Docker containers across a Kubernetes cluster).

## The Activity Registry
A core philosophy of Workflow Engines is separating the "engine code" from the "business code". 

In Chronos, you write standard Go functions (Activities) and register them with the Worker using `w.RegisterActivity()`.

When a Worker pops a task from Redis with the type `"charge_customer"`, it dynamically looks up your registered Go function and executes it.

This means you can easily add hundreds of different background jobs to the system without ever touching the core Chronos engine logic!
