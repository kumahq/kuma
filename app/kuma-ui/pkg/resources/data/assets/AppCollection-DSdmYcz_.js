import{m as q}from"./kong-icons.es245-BjB891cP.js";import{d as B,z as F,N as L,G as u,c as V,R as h,i as D,o as i,a as g,P as C,w as l,j as T,a8 as E,k as r,t as d,A as m,b as S,H as w,e as N,O as H,J as K,h as $,n as W,ab as X,_ as G}from"./index-CyAtMQ3G.js";const J={key:0,class:"app-collection-toolbar"},A=5,U=B({__name:"AppCollection",props:{isSelectedRow:{type:[Function,null],default:null},total:{default:0},pageNumber:{default:0},pageSize:{default:30},items:{},headers:{},error:{default:void 0},emptyStateTitle:{default:void 0},emptyStateMessage:{default:void 0},emptyStateCtaTo:{default:void 0},emptyStateCtaText:{default:void 0}},emits:["change"],setup(x,{emit:z}){const{t:R}=F(),e=x,M=z,O=L(),v=u(e.items),b=u(0),f=u(0),y=u(e.pageNumber),_=u(e.pageSize),I=V(()=>{const t=e.headers.filter(n=>["details","warnings","actions"].includes(n.key));if(t.length>4)return"initial";const a=100-t.length*A,o=e.headers.length-t.length;return`calc(${a}% / ${o})`});h(()=>e.items,(t,a)=>{t!==a&&(b.value++,v.value=e.items)}),h(()=>e.pageNumber,function(){e.pageNumber!==y.value&&f.value++}),h(()=>e.headers,function(){f.value++});function P(t){if(!t)return{};const a={};return e.isSelectedRow!==null&&e.isSelectedRow(t)&&(a.class="is-selected"),a}const j=t=>{const a=t.target.closest("tr");if(a){const o=["td:first-child a","[data-action]"].reduce((n,s)=>n===null?a.querySelector(s):n,null);o!==null&&o.closest("tr, li")===a&&o.click()}};return(t,a)=>{var n;const o=D("XAction");return i(),g(m(X),{key:f.value,class:"app-collection",style:W(`--column-width: ${I.value}; --special-column-width: ${A}%;`),"has-error":typeof e.error<"u","pagination-total-items":e.total,"initial-fetcher-params":{page:e.pageNumber,pageSize:e.pageSize},headers:e.headers,"fetcher-cache-key":String(b.value),fetcher:({page:s,pageSize:c,query:k})=>{const p={};return y.value!==s&&(p.page=s),_.value!==c&&(p.size=c),y.value=s,_.value=c,Object.keys(p).length>0&&M("change",p),{data:v.value}},"cell-attrs":({headerKey:s})=>({class:`${s}-column`}),"row-attrs":P,"disable-sorting":"","disable-pagination":e.pageNumber===0,"hide-pagination-when-optional":"","onRow:click":j},C({_:2},[((n=e.items)==null?void 0:n.length)===0?{name:"empty-state",fn:l(()=>[T(E,null,C({title:l(()=>[r(d(e.emptyStateTitle??m(R)("common.emptyState.title")),1)]),default:l(()=>[r(),e.emptyStateMessage?(i(),S(w,{key:0},[r(d(e.emptyStateMessage),1)],64)):N("",!0),r()]),_:2},[e.emptyStateCtaTo?{name:"action",fn:l(()=>[typeof e.emptyStateCtaTo=="string"?(i(),g(o,{key:0,type:"docs",href:e.emptyStateCtaTo},{default:l(()=>[r(d(e.emptyStateCtaText),1)]),_:1},8,["href"])):(i(),g(m(H),{key:1,appearance:"primary",to:e.emptyStateCtaTo},{default:l(()=>[T(m(q)),r(" "+d(e.emptyStateCtaText),1)]),_:1},8,["to"]))]),key:"0"}:void 0]),1024)]),key:"0"}:void 0,K(Object.keys(m(O)),s=>({name:s,fn:l(({row:c,rowValue:k})=>[s==="toolbar"?(i(),S("div",J,[$(t.$slots,"toolbar",{},void 0,!0)])):(i(),S(w,{key:1},[(e.items??[]).length>0?$(t.$slots,s,{key:0,row:c,rowValue:k},void 0,!0):N("",!0)],64))])}))]),1032,["style","has-error","pagination-total-items","initial-fetcher-params","headers","fetcher-cache-key","fetcher","cell-attrs","disable-pagination"])}}}),Z=G(U,[["__scopeId","data-v-11b03fb4"]]);export{Z as A};
