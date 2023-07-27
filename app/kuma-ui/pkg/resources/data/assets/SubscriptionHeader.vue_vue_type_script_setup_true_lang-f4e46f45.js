import{d as A,ag as j,j as $,c as B,o as n,e as l,q as h,y as x,g as i,h as b,w as d,f as p,T as C,n as E,ah as q,a as f,b as c,t as u,F as k,s as S,v as N,M as V,p as H,m as M}from"./index-065c0e80.js";import{f as O,g as P,r as D,K as L}from"./RouteView.vue_vue_type_script_setup_true_lang-1d679e8a.js";import{a as y,D as w}from"./DefinitionListItem-adce3dd6.js";const F=["aria-expanded"],R={key:0,class:"accordion-item-content","data-testid":"accordion-item-content"},z=A({__name:"AccordionItem",setup(t){const e=j("parentAccordion"),a=$(null),r=B(()=>e===void 0?!1:e.multipleOpen&&Array.isArray(e.active.value)&&a.value!==null?e.active.value.includes(a.value):a.value===e.active.value);e!==void 0&&(a.value=e.count.value++);function _(){r.value?I():o()}function I(){e!==void 0&&(e.multipleOpen&&Array.isArray(e.active.value)&&a.value!==null?e.active.value.splice(e.active.value.indexOf(a.value),1):e.active.value=null)}function o(){e!==void 0&&(e.multipleOpen&&Array.isArray(e.active.value)&&a.value!==null?e.active.value.push(a.value):e.active.value=a.value)}function v(s){s instanceof HTMLElement&&(s.style.height=`${s.scrollHeight}px`)}function m(s){s instanceof HTMLElement&&(s.style.height="auto")}return(s,T)=>(n(),l("li",{class:E(["accordion-item",{active:r.value}])},[h("button",{class:"accordion-item-header",type:"button","aria-expanded":r.value?"true":"false","data-testid":"accordion-item-button",onClick:_},[x(s.$slots,"accordion-header",{},void 0,!0)],8,F),i(),b(C,{name:"accordion",onEnter:v,onAfterEnter:m,onBeforeLeave:v},{default:d(()=>[r.value?(n(),l("div",R,[x(s.$slots,"accordion-content",{},void 0,!0)])):p("",!0)]),_:3})],2))}});const oe=O(z,[["__scopeId","data-v-dfd99690"]]),G={class:"accordion-list"},K=A({__name:"AccordionList",props:{initiallyOpen:{type:[Number,Array],required:!1,default:null},multipleOpen:{type:Boolean,required:!1,default:!1}},setup(t){const e=t,a=$(0),r=$(e.initiallyOpen!==null?e.initiallyOpen:e.multipleOpen?[]:null);return q("parentAccordion",{multipleOpen:e.multipleOpen,active:r,count:a}),(_,I)=>(n(),l("ul",G,[x(_.$slots,"default",{},void 0,!0)]))}});const ce=O(K,[["__scopeId","data-v-53d92d22"]]),U=t=>(H("data-v-321555ca"),t=t(),M(),t),J={key:0},Q=U(()=>h("h5",{class:"overview-tertiary-title"},`
        General Information:
      `,-1)),W={key:1,class:"columns mt-4",style:{"--columns":"4"}},X={key:0},Y={class:"overview-tertiary-title"},Z=A({__name:"SubscriptionDetails",props:{details:{type:Object,required:!0},isDiscoverySubscription:{type:Boolean,default:!1}},setup(t){const e=t,{t:a}=P(),r=B(()=>{var v,m;let o;if(e.isDiscoverySubscription){const{lastUpdateTime:s,total:T,...g}=e.details.status;o=g}return(v=e.details.status)!=null&&v.stat&&(o=(m=e.details.status)==null?void 0:m.stat),o});function _(o){return o?parseInt(o,10).toLocaleString("en").toString():"0"}function I(o){return o==="--"?"error calculating":o}return(o,v)=>(n(),l("div",null,[t.details.globalInstanceId||t.details.controlPlaneInstanceId||t.details.connectTime||t.details.disconnectTime?(n(),l("div",J,[Q,i(),b(w,null,{default:d(()=>[t.details.globalInstanceId?(n(),f(y,{key:0,term:c(a)("http.api.property.globalInstanceId")},{default:d(()=>[i(u(t.details.globalInstanceId),1)]),_:1},8,["term"])):p("",!0),i(),t.details.controlPlaneInstanceId?(n(),f(y,{key:1,term:c(a)("http.api.property.controlPlaneInstanceId")},{default:d(()=>[i(u(t.details.controlPlaneInstanceId),1)]),_:1},8,["term"])):p("",!0),i(),t.details.connectTime?(n(),f(y,{key:2,term:c(a)("http.api.property.connectTime")},{default:d(()=>[i(u(c(D)(t.details.connectTime)),1)]),_:1},8,["term"])):p("",!0),i(),t.details.disconnectTime?(n(),f(y,{key:3,term:c(a)("http.api.property.disconnectTime")},{default:d(()=>[i(u(c(D)(t.details.disconnectTime)),1)]),_:1},8,["term"])):p("",!0)]),_:1})])):p("",!0),i(),r.value?(n(),l("div",W,[(n(!0),l(k,null,S(r.value,(m,s)=>(n(),l(k,{key:s},[Object.keys(m).length>0?(n(),l("div",X,[h("h6",Y,u(c(a)(`http.api.property.${s}`))+`:
          `,1),i(),b(w,null,{default:d(()=>[(n(!0),l(k,null,S(m,(T,g)=>(n(),f(y,{key:g,term:c(a)(`http.api.property.${g}`)},{default:d(()=>[i(u(I(_(T))),1)]),_:2},1032,["term"]))),128))]),_:2},1024)])):p("",!0)],64))),128))])):(n(),f(c(V),{key:2,appearance:"info",class:"mt-4"},{alertIcon:d(()=>[b(c(N),{icon:"portal"})]),alertMessage:d(()=>[i(`
        There are no subscription statistics for `),h("strong",null,u(t.details.id),1)]),_:1}))]))}});const le=O(Z,[["__scopeId","data-v-321555ca"]]),ee={class:"text-lg font-medium"},te={class:"color-green-500"},ae={key:0,class:"ml-4 color-red-600"},re=A({__name:"SubscriptionHeader",props:{details:{type:Object,required:!0}},setup(t){const e=t;return(a,r)=>(n(),l("h4",ee,[h("span",te,`
      Connect time: `+u(c(L)(e.details.connectTime)),1),i(),e.details.disconnectTime?(n(),l("span",ae,`
      Disconnect time: `+u(c(L)(e.details.disconnectTime)),1)):p("",!0)]))}});export{oe as A,le as S,re as _,ce as a};
