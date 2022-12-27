import {ManagedServiceTypeEnum} from "@/api/generated";
import {h} from "vue";

export const types = [
  {
    type: ManagedServiceTypeEnum.Postgres,
    name: 'PostgreSQL 13',
    image: () => h('i', { class: 'bi bi-database' })
  },
  {
    type: ManagedServiceTypeEnum.Mysql,
    name: 'MySQL 8',
    image: () => h('i', { class: 'bi bi-database' })
  },
  {
    type: ManagedServiceTypeEnum.Redis,
    name: 'Redis 7',
    image: () => h('i', { class: 'bi bi-stack' })
  },
  {
    type: ManagedServiceTypeEnum.Rabbitmq,
    name: 'RabbitMQ 3',
    image: () => h('i', { class: 'bi bi-chat-left-dots' })
  }
]

export const TypeImage = {
  props: ['type', 'fontSize'],
  render() {
    const el = ((this as any).type as { type: string, name: string, image: Function}).image()
    el.props.class += ' fs-' + (this as any).fontSize ?? ''
    return el
  }
}
