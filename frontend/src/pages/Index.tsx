import { Link } from "react-router-dom";
import { MessageCircle, Sparkles, Shield, Zap } from "lucide-react";
import { motion } from "framer-motion";

const Index = () => {
  return (
    <div className="flex min-h-screen flex-col bg-mesh">
      {/* Ambient orbs */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute -left-40 -top-40 h-[28rem] w-[28rem] rounded-full bg-primary/15 blur-3xl" />
        <div className="absolute -bottom-40 -right-40 h-96 w-96 rounded-full bg-accent/15 blur-3xl" />
      </div>

      <header className="relative z-10 flex items-center justify-between px-6 py-4 md:px-10">
        <div className="flex items-center gap-2">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl gradient-primary shadow-lg">
            <MessageCircle className="h-5 w-5 text-white" />
          </div>
          <span className="gradient-text text-xl font-bold tracking-tight">Aura</span>
        </div>
        <nav className="flex items-center gap-3">
          <Link
            to="/login"
            className="rounded-xl px-4 py-2 text-sm font-medium text-foreground/80 transition-colors hover:text-foreground hover:bg-white/5"
          >
            Sign in
          </Link>
          <Link
            to="/signup"
            className="rounded-xl gradient-primary px-4 py-2 text-sm font-semibold text-white shadow-lg transition-all hover:shadow-xl hover:opacity-95"
          >
            Get started
          </Link>
        </nav>
      </header>

      <main className="relative z-10 flex flex-1 flex-col items-center justify-center px-6 py-16 text-center md:py-24">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
          className="max-w-3xl"
        >
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/5 px-4 py-1.5 text-sm font-medium text-primary">
            <Sparkles className="h-4 w-4" />
            Modern chat, reimagined
          </div>
          <h1 className="text-4xl font-extrabold tracking-tight text-foreground sm:text-5xl md:text-6xl">
            Stay connected with{" "}
            <span className="gradient-text">conversations that matter</span>
          </h1>
          <p className="mt-6 text-lg text-muted-foreground sm:text-xl">
            Real-time messaging, knowledge base, and collaborative whiteboards — all in one place.
          </p>
          <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
            <Link
              to="/signup"
              className="inline-flex items-center gap-2 rounded-2xl gradient-primary px-8 py-4 text-base font-semibold text-white shadow-xl transition-all hover:shadow-2xl hover:opacity-95"
            >
              Create free account
              <Zap className="h-5 w-5" />
            </Link>
            <Link
              to="/login"
              className="inline-flex items-center gap-2 rounded-2xl border border-border bg-card/50 px-8 py-4 text-base font-medium text-foreground backdrop-blur-sm transition-colors hover:bg-card"
            >
              Sign in
            </Link>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.3, duration: 0.5 }}
          className="mt-16 grid w-full max-w-2xl grid-cols-1 gap-6 sm:grid-cols-3"
        >
          {[
            { icon: MessageCircle, title: "Real-time chat", desc: "Instant messages with presence and reactions" },
            { icon: Shield, title: "Private & secure", desc: "Your data stays yours" },
            { icon: Zap, title: "Fast & reliable", desc: "Built for speed and scale" },
          ].map((item, i) => (
            <div
              key={i}
              className="rounded-2xl border border-border/60 bg-card/40 p-6 text-left backdrop-blur-sm transition-colors hover:border-primary/20 hover:bg-card/60"
            >
              <item.icon className="h-8 w-8 text-primary" />
              <h3 className="mt-3 font-semibold text-foreground">{item.title}</h3>
              <p className="mt-1 text-sm text-muted-foreground">{item.desc}</p>
            </div>
          ))}
        </motion.div>
      </main>
    </div>
  );
};

export default Index;
