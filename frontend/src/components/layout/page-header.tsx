type PageHeaderProps = {
  eyebrow?: string;
  title: string;
  description: string;
};

export function PageHeader({ eyebrow, title, description }: PageHeaderProps) {
  return (
    <header className="max-w-3xl space-y-2">
      {eyebrow ? (
        <p className="text-sm font-medium uppercase tracking-normal text-muted-foreground">
          {eyebrow}
        </p>
      ) : null}
      <h1 className="text-3xl font-semibold tracking-normal text-foreground sm:text-4xl">
        {title}
      </h1>
      <p className="text-base leading-7 text-muted-foreground">{description}</p>
    </header>
  );
}
