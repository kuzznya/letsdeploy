import { ManagedServiceTypeEnum } from "@/api/generated";
import { defineComponent, h, PropType, VNode } from "vue";

export type ManagedServiceType = {
  type: ManagedServiceTypeEnum;
  name: string;
  image: () => VNode;
};

export const types: { [type in ManagedServiceTypeEnum]: ManagedServiceType } = {
  [ManagedServiceTypeEnum.Postgres]: {
    type: ManagedServiceTypeEnum.Postgres,
    name: "PostgreSQL 13",
    image: () => h("i", { class: "bi bi-database" }),
  },
  [ManagedServiceTypeEnum.Mysql]: {
    type: ManagedServiceTypeEnum.Mysql,
    name: "MySQL 8",
    image: () => h("i", { class: "bi bi-database" }),
  },
  [ManagedServiceTypeEnum.Mongo]: {
    type: ManagedServiceTypeEnum.Mongo,
    name: "MongoDB 5",
    image: () => h("i", { class: "bi bi-database" }),
  },
  [ManagedServiceTypeEnum.Redis]: {
    type: ManagedServiceTypeEnum.Redis,
    name: "Redis 7",
    image: () => h("i", { class: "bi bi-stack" }),
  },
  [ManagedServiceTypeEnum.Rabbitmq]: {
    type: ManagedServiceTypeEnum.Rabbitmq,
    name: "RabbitMQ 3",
    image: () => h("i", { class: "bi bi-chat-left-dots" }),
  },
};

export const TypeImage = defineComponent({
  props: {
    type: {
      type: Object as PropType<ManagedServiceType>,
      required: true,
    },
    fontSize: Number,
  },
  setup(props) {
    return () => {
      const el = props.type.image();
      if (el.props != null && props.fontSize)
        el.props.class += " fs-" + props.fontSize;
      return el;
    };
  },
});
