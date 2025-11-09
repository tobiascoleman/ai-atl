"use client";

import { useState } from "react";
import { Send, Bot, User } from "lucide-react";
import { chatbotAPI } from "@/lib/api/chatbot";
import { ChatMessage } from "@/types/api";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

export default function ChatPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const userMessage: ChatMessage = {
      question: input,
      response: "",
      timestamp: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput("");
    setLoading(true);

    try {
      const response = await chatbotAPI.ask(input);
      setMessages((prev) => [
        ...prev.slice(0, -1),
        { ...userMessage, response: response.response },
      ]);
    } catch (error) {
      console.error("Chat error:", error);
      setMessages((prev) => [
        ...prev.slice(0, -1),
        {
          ...userMessage,
          response: "Sorry, I encountered an error. Please try again.",
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  const quickQuestions = [
    "Who should I start at RB this week?",
    "What are the best waiver wire pickups?",
    "Analyze my trade: give CMC, get Justin Jefferson",
    "Who benefits most from Christian McCaffrey's injury?",
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">
          AI Fantasy Assistant
        </h1>
        <p className="text-gray-600 mt-2">
          Get personalized fantasy advice powered by AI
        </p>
      </div>

      <div
        className="bg-white rounded-xl shadow-lg flex flex-col"
        style={{ height: "calc(100vh - 300px)" }}
      >
        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-6 space-y-4">
          {messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-center">
              <Bot className="w-16 h-16 text-blue-600 mb-4" />
              <h2 className="text-xl font-bold text-gray-900 mb-2">
                Ask me anything about fantasy football
              </h2>
              <p className="text-gray-600 mb-6">
                I can help with start/sit decisions, waiver pickups, trades, and
                more
              </p>

              {/* Quick Questions */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 w-full max-w-2xl">
                {quickQuestions.map((question, i) => (
                  <button
                    key={i}
                    onClick={() => setInput(question)}
                    className="p-4 text-left text-base font-medium text-gray-900 bg-gradient-to-br from-blue-50 to-blue-100 hover:from-blue-100 hover:to-blue-200 border border-blue-200 rounded-lg transition-all shadow-sm hover:shadow-md"
                  >
                    <span className="text-blue-600 mr-2">💬</span>
                    {question}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            messages.map((msg, i) => (
              <div key={i}>
                {/* User Message */}
                <div className="flex items-start gap-3 justify-end mb-4">
                  <div className="bg-blue-600 text-white rounded-lg px-4 py-2 max-w-[70%]">
                    {msg.question}
                  </div>
                  <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
                    <User size={16} className="text-blue-600" />
                  </div>
                </div>

                {/* AI Response */}
                {msg.response && (
                  <div className="flex items-start gap-3 mb-4">
                    <div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center flex-shrink-0">
                      <Bot size={16} className="text-purple-600" />
                    </div>
                    <div className="bg-gray-100 rounded-lg px-4 py-2 max-w-[70%] prose prose-sm max-w-none">
                      <ReactMarkdown
                        remarkPlugins={[remarkGfm]}
                        components={{
                          // Style markdown elements
                          h1: ({ node, ...props }) => (
                            <h1
                              className="text-xl font-bold mb-2 text-gray-900"
                              {...props}
                            />
                          ),
                          h2: ({ node, ...props }) => (
                            <h2
                              className="text-lg font-bold mb-2 text-gray-900"
                              {...props}
                            />
                          ),
                          h3: ({ node, ...props }) => (
                            <h3
                              className="text-base font-bold mb-1 text-gray-900"
                              {...props}
                            />
                          ),
                          p: ({ node, ...props }) => (
                            <p className="mb-2 text-gray-800" {...props} />
                          ),
                          ul: ({ node, ...props }) => (
                            <ul
                              className="list-disc list-inside mb-2 space-y-1"
                              {...props}
                            />
                          ),
                          ol: ({ node, ...props }) => (
                            <ol
                              className="list-decimal list-inside mb-2 space-y-1"
                              {...props}
                            />
                          ),
                          li: ({ node, ...props }) => (
                            <li className="text-gray-800" {...props} />
                          ),
                          strong: ({ node, ...props }) => (
                            <strong
                              className="font-bold text-gray-900"
                              {...props}
                            />
                          ),
                          em: ({ node, ...props }) => (
                            <em className="italic" {...props} />
                          ),
                          code: ({ node, inline, ...props }: any) =>
                            inline ? (
                              <code
                                className="bg-gray-200 px-1 py-0.5 rounded text-sm font-mono"
                                {...props}
                              />
                            ) : (
                              <code
                                className="block bg-gray-200 p-2 rounded text-sm font-mono overflow-x-auto"
                                {...props}
                              />
                            ),
                        }}
                      >
                        {msg.response}
                      </ReactMarkdown>
                    </div>
                  </div>
                )}
              </div>
            ))
          )}

          {loading && (
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center">
                <Bot size={16} className="text-purple-600" />
              </div>
              <div className="bg-gray-100 rounded-lg px-4 py-2">
                <div className="flex gap-1">
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                  <div
                    className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                    style={{ animationDelay: "0.1s" }}
                  ></div>
                  <div
                    className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                    style={{ animationDelay: "0.2s" }}
                  ></div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Input */}
        <div className="border-t p-4">
          <form onSubmit={handleSubmit} className="flex gap-3">
            <input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Ask about your lineup, matchups, waiver targets..."
              className="flex-1 px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
              disabled={loading}
            />
            <button
              type="submit"
              disabled={loading || !input.trim()}
              className="px-6 py-3 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition flex items-center gap-2"
            >
              <Send size={18} />
              Send
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
