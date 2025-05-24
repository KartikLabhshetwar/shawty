"use client"

import type React from "react"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card } from "@/components/ui/card"
import { Copy, Check, Link2, Zap, Shield, BarChart3 } from "lucide-react"
import Link from "next/link"

export default function HomePage() {
  const [url, setUrl] = useState("")
  const [shortenedUrl, setShortenedUrl] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null) // Added error state

  const handleShorten = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!url) return

    setIsLoading(true)
    setShortenedUrl("") // Clear previous shortened URL
    setError(null) // Clear previous error

    const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080/shorten";
    const shortenEndpoint = `${apiBaseUrl}/shorten`;

    try {
      const response = await fetch(shortenEndpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ url: url }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({})) // Try to parse error, default to empty object
        throw new Error(
          errorData.message || `Error: ${response.status} ${response.statusText || "Failed to shorten URL"}`,
        )
      }

      const data = await response.json()
      if (data.short_url) {
        setShortenedUrl(data.short_url)
      } else {
        throw new Error("Shortened URL not found in response")
      }
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message)
        console.error("Failed to shorten URL:", err.message)
      } else {
        setError("An unknown error occurred.")
        console.error("Failed to shorten URL:", err)
      }
    } finally {
      setIsLoading(false)
    }
  }

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(shortenedUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      {/* Header */}
      <header className="container mx-auto px-4 py-6">
        <nav className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Link2 className="h-8 w-8 text-slate-900" />
            <span className="text-2xl font-bold text-slate-900">Shawty</span>
          </div>
          <div className="hidden md:flex items-center space-x-6">
            <Link href="#features" className="text-slate-600 hover:text-slate-900 transition-colors">
              Features
            </Link>
            <Link href="#pricing" className="text-slate-600 hover:text-slate-900 transition-colors">
              Pricing
            </Link>
            <Button variant="outline" size="sm">
              Sign In
            </Button>
          </div>
        </nav>
      </header>

      {/* Hero Section */}
      <main className="container mx-auto px-4 py-16 md:py-24">
        <div className="max-w-4xl mx-auto text-center">
          <h1 className="text-5xl md:text-7xl font-bold text-slate-900 mb-6 leading-tight">
            Make your links{" "}
            <span className="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">Shawty</span>
          </h1>
          <p className="text-xl md:text-2xl text-slate-600 mb-12 max-w-2xl mx-auto">
            Give your URLs some swagger. Transform long, messy links into clean, shareable ones.
          </p>

          {/* URL Shortener Form */}
          <Card className="p-8 max-w-2xl mx-auto mb-16 shadow-xl border-0 bg-white/80 backdrop-blur-sm">
            <form onSubmit={handleShorten} className="space-y-6">
              <div className="flex flex-col md:flex-row gap-4">
                <Input
                  type="url"
                  placeholder="Paste your long URL here..."
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  className="flex-1 h-14 text-lg border-slate-200 focus:border-blue-500 focus:ring-blue-500"
                  required
                />
                <Button
                  type="submit"
                  disabled={isLoading || !url}
                  className="h-14 px-8 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-semibold"
                >
                  {isLoading ? (
                    <div className="flex items-center space-x-2">
                      <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                      <span>Shortening...</span>
                    </div>
                  ) : (
                    "Shorten It"
                  )}
                </Button>
              </div>
            </form>

            {/* Shortened URL Result */}
            {shortenedUrl && !error && (
              <div className="mt-8 p-6 bg-slate-50 rounded-lg border">
                <p className="text-sm text-slate-600 mb-2">Your Shawty link:</p>
                <div className="flex items-center justify-between bg-white p-4 rounded-lg border">
                  <code className="text-blue-600 font-mono text-lg flex-1 mr-4">{shortenedUrl}</code>
                  <Button onClick={copyToClipboard} variant="outline" size="sm" className="flex items-center space-x-2">
                    {copied ? (
                      <>
                        <Check className="h-4 w-4" />
                        <span>Copied!</span>
                      </>
                    ) : (
                      <>
                        <Copy className="h-4 w-4" />
                        <span>Copy</span>
                      </>
                    )}
                  </Button>
                </div>
              </div>
            )}

            {/* Error Message */}
            {error && (
              <div className="mt-8 p-4 bg-red-50 text-red-700 rounded-lg border border-red-200">
                <p className="font-semibold">Oops! Something went wrong:</p>
                <p className="text-sm">{error}</p>
              </div>
            )}
          </Card>

          {/* Trust Indicators */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-3xl mx-auto">
            <div className="text-center">
              <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Zap className="h-6 w-6 text-blue-600" />
              </div>
              <h3 className="font-semibold text-slate-900 mb-2">Lightning Fast</h3>
              <p className="text-slate-600">Instant URL shortening with global CDN delivery</p>
            </div>
            <div className="text-center">
              <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Shield className="h-6 w-6 text-green-600" />
              </div>
              <h3 className="font-semibold text-slate-900 mb-2">Secure & Reliable</h3>
              <p className="text-slate-600">Enterprise-grade security with 99.9% uptime</p>
            </div>
            <div className="text-center">
              <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <BarChart3 className="h-6 w-6 text-purple-600" />
              </div>
              <h3 className="font-semibold text-slate-900 mb-2">Detailed Analytics</h3>
              <p className="text-slate-600">Track clicks, locations, and engagement metrics</p>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-200 bg-white/50 backdrop-blur-sm">
        <div className="container mx-auto px-4 py-8">
          <div className="flex flex-col md:flex-row items-center justify-between">
            <div className="flex items-center space-x-2 mb-4 md:mb-0">
              <Link2 className="h-6 w-6 text-slate-600" />
              <span className="text-lg font-semibold text-slate-900">Shawty</span>
            </div>
            <div className="flex items-center space-x-6 text-sm text-slate-600">
              <Link href="/privacy" className="hover:text-slate-900 transition-colors">
                Privacy
              </Link>
              <Link href="/terms" className="hover:text-slate-900 transition-colors">
                Terms
              </Link>
              <Link href="/contact" className="hover:text-slate-900 transition-colors">
                Contact
              </Link>
            </div>
          </div>
          <div className="mt-6 pt-6 border-t border-slate-200 text-center text-sm text-slate-500">
            © {new Date().getFullYear()} Shawty. All rights reserved.
          </div>
        </div>
      </footer>
    </div>
  )
}
