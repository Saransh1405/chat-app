import { Link } from "react-router-dom";
import { Home, MessageCircle } from "lucide-react";
import { motion } from "framer-motion";

const NotFound = () => {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-mesh px-6">
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="absolute -left-40 -top-40 h-96 w-96 rounded-full bg-primary/10 blur-3xl" />
        <div className="absolute -bottom-40 -right-40 h-96 w-96 rounded-full bg-accent/10 blur-3xl" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
        className="relative z-10 w-full max-w-md rounded-3xl border border-border/60 bg-card/60 p-10 text-center shadow-xl backdrop-blur-xl"
      >
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 text-primary">
          <MessageCircle className="h-8 w-8" />
        </div>
        <h1 className="text-6xl font-extrabold tracking-tight text-foreground">404</h1>
        <p className="mt-3 text-lg text-muted-foreground">This page doesn’t exist or was moved.</p>
        <Link
          to="/"
          className="mt-8 inline-flex items-center gap-2 rounded-xl gradient-primary px-6 py-3 text-sm font-semibold text-white shadow-lg transition-all hover:shadow-xl"
        >
          <Home className="h-4 w-4" />
          Back to Aura
        </Link>
      </motion.div>
    </div>
  );
};

export default NotFound;
