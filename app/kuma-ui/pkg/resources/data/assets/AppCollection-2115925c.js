import{d as z,l as I,o as u,c as A,e as g,q as o,K as D,aa as B,f as c,p as K,r as v,t as p,_ as L,Y as j,m as r,R as E,ab as w,b as h,a5 as N,w as l,x as P,X as V,D as U,ac as W,a8 as X}from"./index-646486ee.js";import{_ as F}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-f5640fed.js";const H=["href"],Y=z({__name:"DocumentationLink",props:{href:{}},setup(_){const{t:m}=I(),f=_;return(e,S)=>(u(),A("a",{class:"docs-link",href:f.href,target:"_blank"},[g(o(B),{size:o(D),title:o(m)("common.documentation")},null,8,["size","title"]),c(),K("span",null,[v(e.$slots,"default",{},()=>[c(p(o(m)("common.documentation")),1)],!0)])],8,H))}});const Z=L(Y,[["__scopeId","data-v-1e7645ce"]]),G={key:0,class:"app-collection-toolbar"},x=5,J=z({__name:"AppCollection",props:{isSelectedRow:{type:[Function,null],default:null},total:{default:0},pageNumber:{default:1},pageSize:{default:30},items:{},headers:{},error:{default:void 0},emptyStateTitle:{default:void 0},emptyStateMessage:{default:void 0},emptyStateCtaTo:{default:void 0},emptyStateCtaText:{default:void 0}},emits:["change"],setup(_,{emit:m}){const{t:f}=I(),e=_,S=m,R=j(),k=r(e.items),C=r(0),b=r(0),y=r(e.pageNumber),T=r(e.pageSize),M=E(()=>{const t=e.headers.filter(a=>["details","warnings","actions"].includes(a.key));if(t.length>4)return"initial";const s=100-t.length*x,n=e.headers.length-t.length;return`calc(${s}% / ${n})`});w(()=>e.items,(t,s)=>{t!==s&&(C.value++,k.value=e.items)}),w(()=>e.pageNumber,function(){e.pageNumber!==y.value&&b.value++});function O(t){if(!t)return{};const s={};return e.isSelectedRow!==null&&e.isSelectedRow(t)&&(s.class="is-selected"),s}const q=t=>{const s=t.target.closest("tr");if(s){const n=s.querySelector("a");n!==null&&n.click()}};return(t,s)=>{var n;return u(),h(o(X),{key:b.value,class:"app-collection",style:W(`--column-width: ${M.value}; --special-column-width: ${x}%;`),"has-error":typeof e.error<"u","pagination-total-items":e.total,"initial-fetcher-params":{page:e.pageNumber,pageSize:e.pageSize},headers:e.headers,"fetcher-cache-key":String(C.value),fetcher:({page:a,pageSize:i,query:$})=>{const d={};return y.value!==a&&(d.page=a),T.value!==i&&(d.size=i),y.value=a,T.value=i,Object.keys(d).length>0&&S("change",d),{data:k.value}},"cell-attrs":({headerKey:a})=>({class:`${a}-column`}),"row-attrs":O,"disable-sorting":"","hide-pagination-when-optional":"","onRow:click":q},N({_:2},[((n=e.items)==null?void 0:n.length)===0?{name:"empty-state",fn:l(()=>[g(F,null,N({default:l(()=>[c(p(e.emptyStateTitle??o(f)("common.emptyState.title"))+" ",1),c()]),_:2},[e.emptyStateMessage?{name:"message",fn:l(()=>[c(p(e.emptyStateMessage),1)]),key:"0"}:void 0,e.emptyStateCtaTo?{name:"cta",fn:l(()=>[typeof e.emptyStateCtaTo=="string"?(u(),h(Z,{key:0,href:e.emptyStateCtaTo},{default:l(()=>[c(p(e.emptyStateCtaText),1)]),_:1},8,["href"])):(u(),h(o(P),{key:1,appearance:"primary",to:e.emptyStateCtaTo},{default:l(()=>[g(o(V),{size:o(D)},null,8,["size"]),c(" "+p(e.emptyStateCtaText),1)]),_:1},8,["to"]))]),key:"1"}:void 0]),1024)]),key:"0"}:void 0,U(Object.keys(o(R)),a=>({name:a,fn:l(({row:i,rowValue:$})=>[a==="toolbar"?(u(),A("div",G,[v(t.$slots,"toolbar",{},void 0,!0)])):v(t.$slots,a,{key:1,row:i,rowValue:$},void 0,!0)])}))]),1032,["style","has-error","pagination-total-items","initial-fetcher-params","headers","fetcher-cache-key","fetcher","cell-attrs"])}}});const te=L(J,[["__scopeId","data-v-1dad27d2"]]);export{te as A,Z as D};
