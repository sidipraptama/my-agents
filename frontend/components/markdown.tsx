import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

export function Markdown({ content }: { content: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
        strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
        em: ({ children }) => <em className="italic">{children}</em>,
        ul: ({ children }) => (
          <ul className="mb-2 ml-4 list-disc space-y-1 last:mb-0">{children}</ul>
        ),
        ol: ({ children }) => (
          <ol className="mb-2 ml-4 list-decimal space-y-1 last:mb-0">{children}</ol>
        ),
        li: ({ children }) => <li>{children}</li>,
        h1: ({ children }) => (
          <h1 className="mb-2 mt-3 text-base font-semibold first:mt-0">{children}</h1>
        ),
        h2: ({ children }) => (
          <h2 className="mb-2 mt-3 text-base font-semibold first:mt-0">{children}</h2>
        ),
        h3: ({ children }) => (
          <h3 className="mb-1.5 mt-2.5 text-sm font-semibold first:mt-0">{children}</h3>
        ),
        a: ({ children, href }) => (
          <a
            href={href}
            target="_blank"
            rel="noreferrer"
            className="underline underline-offset-2 opacity-90 hover:opacity-100"
          >
            {children}
          </a>
        ),
        code: ({ className, children }) => {
          const isBlock = /language-/.test(className ?? "");
          if (isBlock) {
            return <code className={className}>{children}</code>;
          }
          return (
            <code className="rounded bg-black/10 px-1 py-0.5 font-mono text-[0.85em] dark:bg-white/10">
              {children}
            </code>
          );
        },
        pre: ({ children }) => (
          <pre className="scrollbar-thin mb-2 overflow-x-auto rounded-lg bg-black/[0.06] p-3 font-mono text-[13px] leading-relaxed last:mb-0 dark:bg-white/[0.08]">
            {children}
          </pre>
        ),
        blockquote: ({ children }) => (
          <blockquote className="mb-2 border-l-2 border-current/20 pl-3 opacity-80 last:mb-0">
            {children}
          </blockquote>
        ),
        hr: () => <hr className="my-3 border-current/10" />,
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
