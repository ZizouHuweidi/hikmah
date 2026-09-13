export default function Footer() {
  const year = new Date().getFullYear();

  return (
    <footer className="mt-20 border-t border-[#d7dbd2] bg-[#f6f5ef] px-4 pb-14 pt-10">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 text-center sm:flex-row sm:text-left">
        <p className="m-0 text-sm text-[#67746e]">&copy; {year} Sabeel. All rights reserved.</p>
        <p className="m-0 text-sm text-[#67746e]">A clearer path through what you learn.</p>
      </div>
    </footer>
  );
}
