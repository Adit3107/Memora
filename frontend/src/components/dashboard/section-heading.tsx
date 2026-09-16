type SectionHeadingProps = {
  id?: string;
  title: string;
  description?: string;
};

export function SectionHeading({ id, title, description }: SectionHeadingProps) {
  return (
    <div className="space-y-1">
      <h2 className="text-lg font-semibold tracking-normal" id={id}>
        {title}
      </h2>
      {description ? (
        <p className="text-sm leading-6 text-muted-foreground">{description}</p>
      ) : null}
    </div>
  );
}
