import{d as B,l as D,L as F,D as u,J as P,ae as h,a as V,o as i,b as g,O as C,w as l,e as T,a9 as j,f as r,t as d,p as m,c as S,F as w,q as N,M as E,S as K,G as W,r as $,af as X,ag as G,_ as H}from"./index-DPw5bDvs.js";const J={key:0,class:"app-collection-toolbar"},A=5,U=B({__name:"AppCollection",props:{isSelectedRow:{type:[Function,null],default:null},total:{default:0},pageNumber:{default:0},pageSize:{default:30},items:{},headers:{},error:{default:void 0},emptyStateTitle:{default:void 0},emptyStateMessage:{default:void 0},emptyStateCtaTo:{default:void 0},emptyStateCtaText:{default:void 0}},emits:["change"],setup(M,{emit:x}){const{t:z}=D(),e=M,O=x,R=F(),v=u(e.items),b=u(0),f=u(0),y=u(e.pageNumber),_=u(e.pageSize),q=P(()=>{const t=e.headers.filter(n=>["details","warnings","actions"].includes(n.key));if(t.length>4)return"initial";const a=100-t.length*A,o=e.headers.length-t.length;return`calc(${a}% / ${o})`});h(()=>e.items,(t,a)=>{t!==a&&(b.value++,v.value=e.items)}),h(()=>e.pageNumber,function(){e.pageNumber!==y.value&&f.value++}),h(()=>e.headers,function(){f.value++});function I(t){if(!t)return{};const a={};return e.isSelectedRow!==null&&e.isSelectedRow(t)&&(a.class="is-selected"),a}const L=t=>{const a=t.target.closest("tr");if(a){const o=["td:first-child a","[data-action]"].reduce((n,s)=>n===null?a.querySelector(s):n,null);o!==null&&o.closest("tr, li")===a&&o.click()}};return(t,a)=>{var n;const o=V("XAction");return i(),g(m(G),{key:f.value,class:"app-collection",style:X(`--column-width: ${q.value}; --special-column-width: ${A}%;`),"has-error":typeof e.error<"u","pagination-total-items":e.total,"initial-fetcher-params":{page:e.pageNumber,pageSize:e.pageSize},headers:e.headers,"fetcher-cache-key":String(b.value),fetcher:({page:s,pageSize:c,query:k})=>{const p={};return y.value!==s&&(p.page=s),_.value!==c&&(p.size=c),y.value=s,_.value=c,Object.keys(p).length>0&&O("change",p),{data:v.value}},"cell-attrs":({headerKey:s})=>({class:`${s}-column`}),"row-attrs":I,"disable-sorting":"","disable-pagination":e.pageNumber===0,"hide-pagination-when-optional":"","onRow:click":L},C({_:2},[((n=e.items)==null?void 0:n.length)===0?{name:"empty-state",fn:l(()=>[T(j,null,C({title:l(()=>[r(d(e.emptyStateTitle??m(z)("common.emptyState.title")),1)]),default:l(()=>[r(),e.emptyStateMessage?(i(),S(w,{key:0},[r(d(e.emptyStateMessage),1)],64)):N("",!0),r()]),_:2},[e.emptyStateCtaTo?{name:"action",fn:l(()=>[typeof e.emptyStateCtaTo=="string"?(i(),g(o,{key:0,type:"docs",href:e.emptyStateCtaTo},{default:l(()=>[r(d(e.emptyStateCtaText),1)]),_:1},8,["href"])):(i(),g(m(E),{key:1,appearance:"primary",to:e.emptyStateCtaTo},{default:l(()=>[T(m(K)),r(" "+d(e.emptyStateCtaText),1)]),_:1},8,["to"]))]),key:"0"}:void 0]),1024)]),key:"0"}:void 0,W(Object.keys(m(R)),s=>({name:s,fn:l(({row:c,rowValue:k})=>[s==="toolbar"?(i(),S("div",J,[$(t.$slots,"toolbar",{},void 0,!0)])):(i(),S(w,{key:1},[(e.items??[]).length>0?$(t.$slots,s,{key:0,row:c,rowValue:k},void 0,!0):N("",!0)],64))])}))]),1032,["style","has-error","pagination-total-items","initial-fetcher-params","headers","fetcher-cache-key","fetcher","cell-attrs","disable-pagination"])}}}),Y=H(U,[["__scopeId","data-v-11b03fb4"]]);export{Y as A};
