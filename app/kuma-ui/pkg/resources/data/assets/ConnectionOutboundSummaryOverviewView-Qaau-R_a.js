import{a as I,A as K}from"./AccordionList-BIwm8-9w.js";import{C as N}from"./CodeBlock-DmzNWuVK.js";import{d as T,r,o as t,m as _,w as e,b as o,c as n,F as d,s as h,X as b,e as a,t as f,k as l,p as g,l as z,G as M,H as j,q as F}from"./index-DxrN05KS.js";import{P as H}from"./PolicyTypeTag-DcpTRUvC.js";import{R as q}from"./RuleMatchers-f3gsQlUX.js";import{t as E}from"./toYaml-DB9FPXFY.js";const B=i=>(M("data-v-9dc2af64"),i=i(),j(),i),G={key:0,class:"rules"},X=B(()=>l("h3",null,"Rules",-1)),Y={class:"mt-4"},J={class:"stack-with-borders"},Q=B(()=>l("dt",null,`
                                    Config
                                  `,-1)),U={class:"mt-2"},W=T({__name:"ConnectionOutboundSummaryOverviewView",props:{data:{},dataplaneOverview:{}},setup(i){const V=i;return(Z,ee)=>{const P=r("KBadge"),R=r("RouterLink"),D=r("DataSource"),O=r("KCard"),S=r("DataCollection"),L=r("DataLoader"),$=r("AppView"),x=r("RouteView");return t(),_(x,{params:{mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-overview-view"},{default:e(({t:A,route:y})=>[o($,null,{default:e(()=>[(t(!0),n(d,null,h([y.params.connection.replace(/-([a-f0-9]){16}$/,"")],v=>(t(),n("div",{key:v,class:"stack-with-borders"},[o(b,{layout:"horizontal"},{title:e(()=>[a(`
              Protocol
            `)]),body:e(()=>[o(P,{appearance:"info"},{default:e(()=>[a(f(A(`http.api.value.${["grpc","http","tcp"].find(k=>typeof V.data[k]<"u")}`)),1)]),_:2},1024)]),_:2},1024),a(),V.data?(t(),n("div",G,[X,a(),o(L,{src:`/meshes/${y.params.mesh}/rules/for/${y.params.dataPlane}`},{default:e(({data:k})=>[o(S,{predicate:p=>p.ruleType==="to"&&!["MeshHTTPRoute","MeshTCPRoute"].includes(p.type)&&p.matchers.every(s=>s.key==="kuma.io/service"&&(s.not?s.value!==v:s.value===v)),items:k.rules},{default:e(({items:p})=>[l("div",Y,[o(I,{"initially-open":0,"multiple-open":"",class:"stack"},{default:e(()=>[(t(!0),n(d,null,h(Object.groupBy(p,s=>s.type),(s,w)=>(t(),_(O,{key:w},{default:e(()=>[o(K,null,{"accordion-header":e(()=>[o(H,{"policy-type":w},{default:e(()=>[a(f(w)+" ("+f(s.length)+`)
                            `,1)]),_:2},1032,["policy-type"])]),"accordion-content":e(()=>[l("div",J,[(t(!0),n(d,null,h(s,u=>(t(),n(d,{key:u},[u.matchers.length>0?(t(),_(b,{key:0,layout:"horizontal"},{title:e(()=>[a(`
                                    To
                                  `)]),body:e(()=>[l("p",null,[o(q,{items:u.matchers},null,8,["items"])])]),_:2},1024)):g("",!0),a(),u.origins.length>0?(t(),_(b,{key:1,layout:"horizontal"},{title:e(()=>[a(`
                                    Origin Policies
                                  `)]),body:e(()=>[o(D,{src:"/policy-types"},{default:e(({data:C})=>[(t(!0),n(d,null,h([Object.groupBy((C==null?void 0:C.policies)??[],m=>m.name)],m=>(t(),n("ul",{key:m},[(t(!0),n(d,null,h(u.origins,c=>(t(),n("li",{key:`${c.mesh}-${c.name}`},[m[c.type]?(t(),_(R,{key:0,to:{name:"policy-detail-view",params:{mesh:c.mesh,policyPath:m[c.type][0].path,policy:c.name}}},{default:e(()=>[a(f(c.name),1)]),_:2},1032,["to"])):(t(),n(d,{key:1},[a(f(c.name),1)],64))]))),128))]))),128))]),_:2},1024)]),_:2},1024)):g("",!0),a(),l("div",null,[Q,a(),l("dd",U,[l("div",null,[o(N,{code:z(E)(u.raw),language:"yaml","show-copy-button":!1},null,8,["code"])])])])],64))),128))])]),_:2},1024)]),_:2},1024))),128))]),_:2},1024)])]),_:2},1032,["predicate","items"])]),_:2},1032,["src"])])):g("",!0)]))),128))]),_:2},1024)]),_:1})}}}),re=F(W,[["__scopeId","data-v-9dc2af64"]]);export{re as default};