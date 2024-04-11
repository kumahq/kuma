import{d as R,l as A,o as c,c as d,e as k,p as l,K as V,a4 as q,f as n,m as j,r as b,t as m,_ as B,Q as E,B as u,J as F,a5 as S,b as v,V as z,w as r,F as I,q as x,R as P,z as U,a0 as W,D as H,a6 as J,a7 as Q}from"./index-C5Gyl9xU.js";const Z=["href"],G=R({__name:"DocumentationLink",props:{href:{}},setup(f){const{t:h}=A(),y=f;return(e,C)=>(c(),d("a",{class:"docs-link",href:y.href,target:"_blank"},[k(l(q),{size:l(V)},null,8,["size"]),n(),j("span",null,[b(e.$slots,"default",{},()=>[n(m(l(h)("common.documentation")),1)],!0)])],8,Z))}}),X=B(G,[["__scopeId","data-v-b5a70c14"]]),Y={key:0,class:"app-collection-toolbar"},D=5,ee=R({__name:"AppCollection",props:{isSelectedRow:{type:[Function,null],default:null},total:{default:0},pageNumber:{default:0},pageSize:{default:30},items:{},headers:{},error:{default:void 0},emptyStateTitle:{default:void 0},emptyStateMessage:{default:void 0},emptyStateCtaTo:{default:void 0},emptyStateCtaText:{default:void 0}},emits:["change"],setup(f,{emit:h}){const{t:y}=A(),e=f,C=h,K=E(),N=u(e.items),T=u(0),g=u(0),_=u(e.pageNumber),$=u(e.pageSize),L=F(()=>{const a=e.headers.filter(t=>["details","warnings","actions"].includes(t.key));if(a.length>4)return"initial";const s=100-a.length*D,o=e.headers.length-a.length;return`calc(${s}% / ${o})`});S(()=>e.items,(a,s)=>{a!==s&&(T.value++,N.value=e.items)}),S(()=>e.pageNumber,function(){e.pageNumber!==_.value&&g.value++}),S(()=>e.headers,function(){g.value++});function M(a){if(!a)return{};const s={};return e.isSelectedRow!==null&&e.isSelectedRow(a)&&(s.class="is-selected"),s}const O=a=>{const s=a.target.closest("tr");if(s){const o=["td:first-child a","[data-action]"].reduce((t,i)=>t===null?s.querySelector(i):t,null);o!==null&&o.closest("tr, li")===s&&o.click()}};return(a,s)=>{var o;return c(),v(l(Q),{key:g.value,class:"app-collection",style:J(`--column-width: ${L.value}; --special-column-width: ${D}%;`),"has-error":typeof e.error<"u","pagination-total-items":e.total,"initial-fetcher-params":{page:e.pageNumber,pageSize:e.pageSize},headers:e.headers,"fetcher-cache-key":String(T.value),fetcher:({page:t,pageSize:i,query:w})=>{const p={};return _.value!==t&&(p.page=t),$.value!==i&&(p.size=i),_.value=t,$.value=i,Object.keys(p).length>0&&C("change",p),{data:N.value}},"cell-attrs":({headerKey:t})=>({class:`${t}-column`}),"row-attrs":M,"disable-sorting":"","disable-pagination":e.pageNumber===0,"hide-pagination-when-optional":"","onRow:click":O},z({_:2},[((o=e.items)==null?void 0:o.length)===0?{name:"empty-state",fn:r(()=>[k(W,null,z({title:r(()=>[n(m(e.emptyStateTitle??l(y)("common.emptyState.title")),1)]),default:r(()=>[n(),e.emptyStateMessage?(c(),d(I,{key:0},[n(m(e.emptyStateMessage),1)],64)):x("",!0),n()]),_:2},[e.emptyStateCtaTo?{name:"action",fn:r(()=>[typeof e.emptyStateCtaTo=="string"?(c(),v(X,{key:0,href:e.emptyStateCtaTo},{default:r(()=>[n(m(e.emptyStateCtaText),1)]),_:1},8,["href"])):(c(),v(l(P),{key:1,appearance:"primary",to:e.emptyStateCtaTo},{default:r(()=>[k(l(U)),n(" "+m(e.emptyStateCtaText),1)]),_:1},8,["to"]))]),key:"0"}:void 0]),1024)]),key:"0"}:void 0,H(Object.keys(l(K)),t=>({name:t,fn:r(({row:i,rowValue:w})=>[t==="toolbar"?(c(),d("div",Y,[b(a.$slots,"toolbar",{},void 0,!0)])):(c(),d(I,{key:1},[(e.items??[]).length>0?b(a.$slots,t,{key:0,row:i,rowValue:w},void 0,!0):x("",!0)],64))])}))]),1032,["style","has-error","pagination-total-items","initial-fetcher-params","headers","fetcher-cache-key","fetcher","cell-attrs","disable-pagination"])}}}),ae=B(ee,[["__scopeId","data-v-765f6ee2"]]);export{ae as A,X as D};
