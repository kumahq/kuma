import{d as v,aj as _,C as o,M as h,o as c,c as l,m as y,r,f as A,e as O,w as b,q as x,ak as C,n as I,_ as p,ag as $}from"./index-BvBYCJa7.js";const k=["aria-expanded"],B={key:0,class:"accordion-item-content","data-testid":"accordion-item-content"},E=v({__name:"AccordionItem",setup(s){const e=_("parentAccordion"),t=o(null),n=h(()=>e===void 0?!1:e.multipleOpen&&Array.isArray(e.active.value)&&t.value!==null?e.active.value.includes(t.value):t.value===e.active.value);e!==void 0&&(t.value=e.count.value++);function i(){n.value?u():f()}function u(){e!==void 0&&(e.multipleOpen&&Array.isArray(e.active.value)&&t.value!==null?e.active.value.splice(e.active.value.indexOf(t.value),1):e.active.value=null)}function f(){e!==void 0&&(e.multipleOpen&&Array.isArray(e.active.value)&&t.value!==null?e.active.value.push(t.value):e.active.value=t.value)}function d(a){a instanceof HTMLElement&&(a.style.height=`${a.scrollHeight}px`)}function m(a){a instanceof HTMLElement&&(a.style.height="auto")}return(a,g)=>(c(),l("li",{class:I(["accordion-item",{active:n.value}])},[y("button",{class:"accordion-item-header",type:"button","aria-expanded":n.value?"true":"false","data-testid":"accordion-item-button",onClick:i},[r(a.$slots,"accordion-header",{},void 0,!0)],8,k),A(),O(C,{name:"accordion",onEnter:d,onAfterEnter:m,onBeforeLeave:d},{default:b(()=>[n.value?(c(),l("div",B,[r(a.$slots,"accordion-content",{},void 0,!0)])):x("",!0)]),_:3})],2))}}),V=p(E,[["__scopeId","data-v-53a0b6ce"]]),L={class:"accordion-list"},N=v({__name:"AccordionList",props:{initiallyOpen:{type:[Number,Array],required:!1,default:null},multipleOpen:{type:Boolean,required:!1,default:!1}},setup(s){const e=s,t=o(0),n=o(e.initiallyOpen!==null?e.initiallyOpen:e.multipleOpen?[]:null);return $("parentAccordion",{multipleOpen:e.multipleOpen,active:n,count:t}),(i,u)=>(c(),l("ul",L,[r(i.$slots,"default",{},void 0,!0)]))}}),q=p(N,[["__scopeId","data-v-bdbadd5e"]]);export{V as A,q as a};
