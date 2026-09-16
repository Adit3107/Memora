import { notFound } from "next/navigation";

import { ContentDetail } from "@/components/content/content-detail";
import { getContentBySlug, savedContent } from "@/data/content";

type ContentDetailPageProps = {
  params: Promise<{
    slug: string;
  }>;
};

export function generateStaticParams() {
  return savedContent.map((item) => ({ slug: item.slug }));
}

export default async function ContentDetailPage({
  params,
}: ContentDetailPageProps) {
  const { slug } = await params;
  const item = getContentBySlug(slug);

  if (!item) {
    notFound();
  }

  return <ContentDetail item={item} />;
}
