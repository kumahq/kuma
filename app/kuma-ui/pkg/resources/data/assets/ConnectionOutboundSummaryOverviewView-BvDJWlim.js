import{a as I,A as K}from"./AccordionList-Bii-UPuH.js";import{C as N}from"./CodeBlock-MXhLppFn.js";import{d as T,a as l,o as t,b as _,w as e,e as o,c as n,F as u,G as h,Z as C,f as a,t as y,m as r,q as g,p as z,x as M,y as j,_ as F}from"./index-DMiSr0Gl.js";import{P as q}from"./PolicyTypeTag-DtfLREOK.js";import{R as E}from"./RuleMatchers-DkjXxSgc.js";import{t as G}from"./toYaml-DB9FPXFY.js";const B=i=>(M("data-v-9280bbb5"),i=i(),j(),i),H={key:0,class:"rules"},Y=B(()=>r("h3",null,"Rules",-1)),Z={class:"stack mt-4"},J={class:"stack-with-borders"},Q=B(()=>r("dt",null,`
                                    Config
                                  `,-1)),U={class:"mt-2"},W=T({__name:"ConnectionOutboundSummaryOverviewView",props:{data:{},dataplaneOverview:{}},setup(i){const V=i;return(X,ee)=>{const P=l("KBadge"),R=l("RouterLink"),D=l("DataSource"),O=l("KCard"),S=l("DataCollection"),x=l("DataLoader"),L=l("AppView"),$=l("RouteView");return t(),_($,{params:{mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-overview-view"},{default:e(({t:A,route:f})=>[o(L,null,{default:e(()=>[(t(!0),n(u,null,h([f.params.connection.replace(/-([a-f0-9]){16}$/,"")],v=>(t(),n("div",{key:v,class:"stack-with-borders"},[o(C,{layout:"horizontal"},{title:e(()=>[a(`
              Protocol
            `)]),body:e(()=>[o(P,{appearance:"info"},{default:e(()=>[a(y(A(`http.api.value.${["grpc","http","tcp"].find(b=>typeof V.data[b]<"u")}`)),1)]),_:2},1024)]),_:2},1024),a(),V.data?(t(),n("div",H,[Y,a(),o(x,{src:`/meshes/${f.params.mesh}/dataplanes/${f.params.dataPlane}/rules`},{default:e(({data:b})=>[o(S,{predicate:p=>p.ruleType==="to"&&!["MeshHTTPRoute","MeshTCPRoute"].includes(p.type)&&p.matchers.every(s=>s.key==="kuma.io/service"&&(s.not?s.value!==v:s.value===v)),items:b.rules},{default:e(({items:p})=>[r("div",Z,[o(I,{"initially-open":0,"multiple-open":""},{default:e(()=>[(t(!0),n(u,null,h(Object.groupBy(p,s=>s.type),(s,k)=>(t(),_(O,{key:k},{default:e(()=>[o(K,null,{"accordion-header":e(()=>[o(q,{"policy-type":k},{default:e(()=>[a(y(k)+" ("+y(s.length)+`)
                            `,1)]),_:2},1032,["policy-type"])]),"accordion-content":e(()=>[r("div",J,[(t(!0),n(u,null,h(s,d=>(t(),n(u,{key:d},[d.matchers.length>0?(t(),_(C,{key:0,layout:"horizontal"},{title:e(()=>[a(`
                                    To
                                  `)]),body:e(()=>[r("p",null,[o(E,{items:d.matchers},null,8,["items"])])]),_:2},1024)):g("",!0),a(),d.origins.length>0?(t(),_(C,{key:1,layout:"horizontal"},{title:e(()=>[a(`
                                    Origin Policies
                                  `)]),body:e(()=>[o(D,{src:"/*/policy-types"},{default:e(({data:w})=>[(t(!0),n(u,null,h([Object.groupBy((w==null?void 0:w.policies)??[],m=>m.name)],m=>(t(),n("ul",{key:m},[(t(!0),n(u,null,h(d.origins,c=>(t(),n("li",{key:`${c.mesh}-${c.name}`},[m[c.type]?(t(),_(R,{key:0,to:{name:"policy-detail-view",params:{mesh:c.mesh,policyPath:m[c.type][0].path,policy:c.name}}},{default:e(()=>[a(y(c.name),1)]),_:2},1032,["to"])):(t(),n(u,{key:1},[a(y(c.name),1)],64))]))),128))]))),128))]),_:2},1024)]),_:2},1024)):g("",!0),a(),r("div",null,[Q,a(),r("dd",U,[r("div",null,[o(N,{code:z(G)(d.config),language:"yaml","show-copy-button":!1},null,8,["code"])])])])],64))),128))])]),_:2},1024)]),_:2},1024))),128))]),_:2},1024)])]),_:2},1032,["predicate","items"])]),_:2},1032,["src"])])):g("",!0)]))),128))]),_:2},1024)]),_:1})}}}),le=F(W,[["__scopeId","data-v-9280bbb5"]]);export{le as default};
