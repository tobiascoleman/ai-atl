import Link from "next/link";
import { Activity, Brain, TrendingUp, Users } from "lucide-react";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Hero Section */}
      <div className="container mx-auto px-4 py-16">
        <div className="text-center mb-16">
          <h1 className="text-6xl font-bold text-gray-900 mb-4">
            AI-Powered NFL Fantasy
          </h1>
          <p className="text-xl text-gray-600 mb-8">
            Predict game flow, optimize lineups, and dominate your league with
            AI insights
          </p>
          <div className="flex gap-4 justify-center">
            <Link
              href="/login"
              className="px-8 py-3 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition"
            >
              Sign In
            </Link>
            <Link
              href="/register"
              className="px-8 py-3 bg-white text-blue-600 rounded-lg font-semibold border-2 border-blue-600 hover:bg-blue-50 transition"
            >
              Register
            </Link>
          </div>
        </div>

        {/* Features Grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8 mt-16">
          <FeatureCard
            icon={<Brain className="w-12 h-12 text-blue-600" />}
            title="AI Game Script Predictor"
            description="Predict how games will unfold quarter by quarter with AI-powered analysis"
          />
          <FeatureCard
            icon={<Activity className="w-12 h-12 text-green-600" />}
            title="EPA-Based Analytics"
            description="Advanced efficiency metrics to find undervalued players"
          />
          <FeatureCard
            icon={<TrendingUp className="w-12 h-12 text-purple-600" />}
            title="Smart Recommendations"
            description="AI chatbot provides personalized lineup and waiver advice"
          />
        </div>
      </div>
    </div>
  );
}

function FeatureCard({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <div className="bg-white rounded-xl p-6 shadow-lg hover:shadow-xl transition">
      <div className="mb-4">{icon}</div>
      <h3 className="text-xl font-bold mb-2 text-gray-900">{title}</h3>
      <p className="text-gray-600">{description}</p>
    </div>
  );
}
