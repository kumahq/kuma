import{K as z}from"./index-fce48c05.js";import{d as I,l as A,o as u,c as D,e as g,q as o,a9 as R,f as c,p as q,r as v,t as p,_ as L,T as K,m as r,M as E,aa as $,b as h,Z as N,w as l,P as V,B as j,I as U,ab as W,ac as Z}from"./index-364a2367.js";import{_ as F}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-3f5123fb.js";const H=["href"],G=I({__name:"DocumentationLink",props:{href:{}},setup(f){const{t:m}=A(),_=f;return(e,S)=>(u(),D("a",{class:"docs-link",href:_.href,target:"_blank"},[g(o(R),{size:o(z),title:o(m)("common.documentation")},null,8,["size","title"]),c(),q("span",null,[v(e.$slots,"default",{},()=>[c(p(o(m)("common.documentation")),1)],!0)])],8,H))}});const J=L(G,[["__scopeId","data-v-1e7645ce"]]),Q={key:0,class:"app-collection-toolbar"},x=5,X=I({__name:"AppCollection",props:{isSelectedRow:{type:[Function,null],default:null},total:{default:0},pageNumber:{default:1},pageSize:{default:30},items:{},headers:{},error:{default:void 0},emptyStateTitle:{default:void 0},emptyStateMessage:{default:void 0},emptyStateCtaTo:{default:void 0},emptyStateCtaText:{default:void 0}},emits:["change"],setup(f,{emit:m}){const{t:_}=A(),e=f,S=m,M=K(),k=r(e.items),C=r(0),b=r(0),y=r(e.pageNumber),T=r(e.pageSize),P=E(()=>{const t=e.headers.filter(a=>["details","warnings","actions"].includes(a.key));if(t.length>4)return"initial";const s=100-t.length*x,n=e.headers.length-t.length;return`calc(${s}% / ${n})`});$(()=>e.items,(t,s)=>{t!==s&&(C.value++,k.value=e.items)}),$(()=>e.pageNumber,function(){e.pageNumber!==y.value&&b.value++});function B(t){if(!t)return{};const s={};return e.isSelectedRow!==null&&e.isSelectedRow(t)&&(s.class="is-selected"),s}const O=t=>{const s=t.target.closest("tr");if(s){const n=s.querySelector("a");n!==null&&n.click()}};return(t,s)=>{var n;return u(),h(o(Z),{key:b.value,class:"app-collection",style:W(`--column-width: ${P.value}; --special-column-width: ${x}%;`),"has-error":typeof e.error<"u","pagination-total-items":e.total,"initial-fetcher-params":{page:e.pageNumber,pageSize:e.pageSize},headers:e.headers,"fetcher-cache-key":String(C.value),fetcher:({page:a,pageSize:i,query:w})=>{const d={};return y.value!==a&&(d.page=a),T.value!==i&&(d.size=i),y.value=a,T.value=i,Object.keys(d).length>0&&S("change",d),{data:k.value}},"cell-attrs":({headerKey:a})=>({class:`${a}-column`}),"row-attrs":B,"disable-sorting":"","hide-pagination-when-optional":"","onRow:click":O},N({_:2},[((n=e.items)==null?void 0:n.length)===0?{name:"empty-state",fn:l(()=>[g(F,null,N({default:l(()=>[c(p(e.emptyStateTitle??o(_)("common.emptyState.title"))+" ",1),c()]),_:2},[e.emptyStateMessage?{name:"message",fn:l(()=>[c(p(e.emptyStateMessage),1)]),key:"0"}:void 0,e.emptyStateCtaTo?{name:"cta",fn:l(()=>[typeof e.emptyStateCtaTo=="string"?(u(),h(J,{key:0,href:e.emptyStateCtaTo},{default:l(()=>[c(p(e.emptyStateCtaText),1)]),_:1},8,["href"])):(u(),h(o(V),{key:1,appearance:"primary",to:e.emptyStateCtaTo},{default:l(()=>[g(o(j),{size:o(z)},null,8,["size"]),c(" "+p(e.emptyStateCtaText),1)]),_:1},8,["to"]))]),key:"1"}:void 0]),1024)]),key:"0"}:void 0,U(Object.keys(o(M)),a=>({name:a,fn:l(({row:i,rowValue:w})=>[a==="toolbar"?(u(),D("div",Q,[v(t.$slots,"toolbar",{},void 0,!0)])):v(t.$slots,a,{key:1,row:i,rowValue:w},void 0,!0)])}))]),1032,["style","has-error","pagination-total-items","initial-fetcher-params","headers","fetcher-cache-key","fetcher","cell-attrs"])}}});const ae=L(X,[["__scopeId","data-v-06f2a961"]]);export{ae as A,J as D};
