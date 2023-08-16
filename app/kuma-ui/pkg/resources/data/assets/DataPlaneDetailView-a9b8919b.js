import{d as Y,r as X,o as e,e as s,g as a,F as v,s as Q,q as h,t as d,h as i,w as t,f as _,a as y,B as Be,b as f,Y as Ce,p as ke,m as Oe,c as x,v as me,z as Ee,$ as he,j as S,I as Pe,R as Ge,G as N,S as Qe,U as Ue}from"./index-9dae1b47.js";import{A as Z,a as $,S as Ie,b as ze}from"./SubscriptionHeader-469a3f33.js";import{f as K,e as ae,m as ve,q as ge,E as fe,r as we,g as Te,D as V,S as Le,A as Me,o as Re,_ as xe}from"./RouteView.vue_vue_type_script_setup_true_lang-f85d915a.js";import{_ as be}from"./CodeBlock.vue_vue_type_style_index_0_lang-f9ef060f.js";import{T as j}from"./TagList-ea500918.js";import{t as ne,_ as _e}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-85a34b5c.js";import{D as Ye,a as Se}from"./DefinitionListItem-73253125.js";import{_ as Ne,E as te}from"./EnvoyData-9b4a9dc8.js";import{T as He}from"./TabsWidget-5339c8e6.js";import{T as qe}from"./TextWithCopyButton-976bd7f2.js";import{_ as je}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-d3bf85b2.js";import{a as Ke,d as ie,b as Fe,p as Je,c as We,C as Ve,I as Xe,e as Ze}from"./dataplane-30467516.js";import{_ as $e}from"./RouteTitle.vue_vue_type_script_setup_true_lang-cc34925e.js";import"./CopyButton-551860fe.js";const F=u=>(ke("data-v-d6898838"),u=u(),Oe(),u),et={class:"mesh-gateway-policy-list"},tt=F(()=>h("h3",{class:"mb-2"},`
      Gateway policies
    `,-1)),at={key:0},st=F(()=>h("h3",{class:"mt-6 mb-2"},`
      Listeners
    `,-1)),nt=F(()=>h("b",null,"Host",-1)),it=F(()=>h("h4",{class:"mt-2 mb-2"},`
              Routes
            `,-1)),ot={class:"dataplane-policy-header"},lt=F(()=>h("b",null,"Route",-1)),At=F(()=>h("b",null,"Service",-1)),rt={key:0,class:"badge-list"},ct={class:"mt-1"},ut=Y({__name:"MeshGatewayDataplanePolicyList",props:{meshGatewayDataplane:{type:Object,required:!0},meshGatewayListenerEntries:{type:Array,required:!0},meshGatewayRoutePolicies:{type:Array,required:!0}},setup(u){const o=u;return(c,O)=>{const D=X("router-link");return e(),s("div",et,[tt,a(),u.meshGatewayRoutePolicies.length>0?(e(),s("ul",at,[(e(!0),s(v,null,Q(u.meshGatewayRoutePolicies,(m,T)=>(e(),s("li",{key:T},[h("span",null,d(m.type),1),a(`:

        `),i(D,{to:m.route},{default:t(()=>[a(d(m.name),1)]),_:2},1032,["to"])]))),128))])):_("",!0),a(),st,a(),h("div",null,[(e(!0),s(v,null,Q(o.meshGatewayListenerEntries,(m,T)=>(e(),s("div",{key:T},[h("div",null,[h("div",null,[nt,a(": "+d(m.hostName)+":"+d(m.port)+" ("+d(m.protocol)+`)
          `,1)]),a(),m.routeEntries.length>0?(e(),s(v,{key:0},[it,a(),i($,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(v,null,Q(m.routeEntries,(A,b)=>(e(),y(Z,{key:b},Be({"accordion-header":t(()=>[h("div",ot,[h("div",null,[h("div",null,[lt,a(": "),i(D,{to:A.route},{default:t(()=>[a(d(A.routeName),1)]),_:2},1032,["to"])]),a(),h("div",null,[At,a(": "+d(A.service),1)])]),a(),A.policies.length>0?(e(),s("div",rt,[(e(!0),s(v,null,Q(A.policies,(n,w)=>(e(),y(f(Ce),{key:`${T}-${w}`},{default:t(()=>[a(d(n.type),1)]),_:2},1024))),128))])):_("",!0)])]),_:2},[A.policies.length>0?{name:"accordion-content",fn:t(()=>[h("ul",ct,[(e(!0),s(v,null,Q(A.policies,(n,w)=>(e(),s("li",{key:`${T}-${w}`},[a(d(n.type)+`:

                      `,1),i(D,{to:n.route},{default:t(()=>[a(d(n.name),1)]),_:2},1032,["to"])]))),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):_("",!0)])]))),128))])])}}});const pt=K(ut,[["__scopeId","data-v-d6898838"]]),oe="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAAB4CAMAAAAOusbgAAAAVFBMVEXa2tra2tra2tra2tra2tra2tr////a2toAfd6izPLvzPnRfvDYteSKr86zas0Aar4AhODY6vr3+Prx8v2Kv+9aqOk3muUOj+N5t+211vXhqfW01fXvn55GAAAABnRSTlMC9s/Hbhsvz/I3AAABVklEQVRo3u3b3Y6CMBCG4SJYhnV/KD+K7v3f57bN7AFJTcDUmZB+74lH5EmMA5hmjK+pq1awqm5M6HxqxTudPSzssmxM06rUmDp8DFawIYi1qYRdlisTeCtcMAGnAgwYMGDAgJ8GGPDB4B8frepnl9cZH5d1374E7GmX1WVuA0xzTvixA+5zwpc0/OXrVgU5N/yx6tMHGDBgwIABvxmeiBZhmF3fPMjDFLuOSjDdnBJMvVOAb1G+y8PjlUKdOGyHOcpLJniiDfEVC/FYZYA3unxFx2OVAd7sTjZ073msRGB2Yy7KvcsC2z05Hitx2P6PVTEwf9W/h/5xvTBOB76ByN8ydzRRzofELln1schjVNCrTxyjsl5vtV7ol7L+tAEGDLhMWOAw5ADHPxIHXmpHfAWepgJOBBgwYMCAAT8NMGDAgJOw2hKO2tqR2qKV1mqZ3jKd2vrgH/W3idgykdWgAAAAAElFTkSuQmCC",dt="data:image/png;base64,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",le="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAAEj0lEQVRoBe1aS28TVxT+7ngydhwnPGISTBKHEB6iUtOWHTvWqCtKqQhISC2vBbCpRDf9BUgsgAWbdlGppWqlSl1UXfMLCmXRqgXUxmCclOCWxI4Tv4Zz7s00BntmLh4rTCSfxJ4755458333fHfuTTQCZFOHTo+ijCs2cAi2nWJfaE2InABuw8Lle7e/eCwYvL2CXwF7a2hBtwQm8iKKdwwe+Y0HnhnRgBN2Q8qmJcPwOxm7EXrNe40jzVfDq38j9HUJvOkqdSvQrUDAEeiIhGaPH8bsyfe1oWQuTuPxhePa8V6BplenTl85tQ2l9A7YMUsnHMsTIyjtm9CK1QkKXIHC3nEI2l3RgqhzPzw/sB/g+A5ZYAKlPTsVFMnCH1Xx3f26XP2TUUQgAuXhJKr9fQqQRgVYPpUtA7IANvQq5sciEIHi7jHKb5OE9DQh5SOvoGs6pKNABJYn06tAaDQ1SLB82DoFnnO1TaA8NIhqIo7IQkFLDI58zPx/WvEMTsfaJlAiPbPF789oiWHxPTX6A3f/kPGdmQEBKlCaGJUE+oiANJ9JvEAEeOL23/ldHvVmjUrt9d1WBSrJLaiRfMzCEqzcU8pPcDzmAMunSk8f699FxP7KqngvVK/R19ZKvDy+Qy5cvQ8z8la2xuhzII8+m9foF9+axOz0YRm3/dbP6PvtoWy7fZm1iIV6tAd1i4+W3BLUrR7Y1Jb+1T7eKqg41ccajj94JPPy4DskaoleZM8cRYmeUGyO1hm0Q6DRz5XMnj2KpV1jTcSYyOTnNzjc1Uw1eCwBpQIFhNWqfvhKCZDPZbCQoGK5eVhz82uJKYjBPDp/DFwhBswZnEcmT3YlnzV/jRbBzKVplFNDTeDXEnu3TLNeBpb44x3o20vksh8fQYU2d1GaF+nr3yBCc6SVOaQyl05gxYm/9rWMf1VCra5v9LU1BxoT/N+mCpSHB2HNzmP05neu4J14ltZKKqnIroLnPta8n2ycHHzsHAGqgPXPM4x8+QOBLzXeo6ntSMsiGaYbwDcFajg6QiA6k0M9EQM/NSJFb/CMqe/PDD0QTKrU976V8uMg3j74ifOg8IsNZX9bC1mYmHQJvOlqBJ7EcUPgw8EELFq5vn1WQKHmPaX6IwIXhzdJ3jfmnmPRJ95vgAJJqJfAf0Tgx3pMpGn7cW5oExIE0M0Y/GepzdgT65EfbrPvVZuKW7g6vlV+uO1lYurgWTtmGHIEo7QYxYhSlM6jlJf9UT6nNvtiBFj5+SjUNeRbrNWpLTBmRSiOc6h8bjfOlquya8TyEQDdN1+t4dOZvFsqXsjU3ob/rqVfMv5iGaijbdORO2ihUlshiqdu5RZ4Uqnix3wRBsWcSiawj/8/xAEqGSd8ye4vV8DS4e3EheEBWYmXAl7zJJTrAMvm1LaEpPLV0wLu8V7NxUJJwAVrS3egSdwy4zo7uwTWecCbbtetQNOQrLPDoOd1bp3v2bnbEXZaN+nFiQ1qjJ3WfFymZdN9rQ4tOcJM2CNzf/+ysH33gVuiLlIkpyTh7Q8tZgbGr9sI8RO9qfIBv27zAiEVYZQrGIvuAAAAAElFTkSuQmCC",Ae="data:image/png;base64,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",re="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAAB5CAYAAADyOOV3AAAFOklEQVR4Ae2dP2tUQRTFp7S385MofgRFiGBhKr9GuqBiI9iktwosCAnRLo0g8Q+ojSaNBomumESTIAqSLiPTTc4k7+bmztudu3sCAztv7p137/nNebtuREMIIXz9eXBluLO/NNzZe8sxCRrsL23tHlxObMP33b3ZzeHO0edv25FjcjRITBPbsPXj12+CnRywOcvENtC9kwk3gU5sQ048vf7775DDsQbIswAc+eNaAQJ2jU8unoBljVxHELBrfHLxBCxr5DqCgF3jk4snYFkj1xEE7BqfXDwByxq5jiBg1/jk4glY1sh1BAG7xicXT8CyRq4jCNg1Prl4ApY1ch1BwK7xycUTsKyR6wgCdo1PLp6AZY1cRxCwa3xy8QQsa+Q6goBd45OLJ2BZI9cRBOwan1w8AcsauY4gYNf45OIJWNbIdQQBu8YnF0/AskauIwjYNT65eAKWNXIdUQD+c2sm5iPemY2mIcnD/bsVMuqTs0yvQ7wQYtXRXb79XtxfpSEB4wH3foCgHwIGQSS+5qeddAOsxzgPOwsLMR9xsBhNQ2qA+3crZNQnZ5le89/o6Jbb3WrxKRovuOuIBR9TAHnSwcfk8T8hYP8MOzsg4E55/C8SsH+GnR0QcKc8/hcJ2D/Dzg4IuFMe/4sE7J9hZwcE3CmP/8WpAzz7cCnm48bdQaw58r3T63H/TB3gcG0+jnIQ8IgVGCXcdK9x/9DBPTuagEesADr43uBFrDlwf217+B5unV+fX4z5mPjfJiGA95vbsebA/bWAMb/6HJ/Z2gJbj0fBasJNe+H+Wj0wv/qcgG2ORiAErFXAGI8AWnfwo5U30TLmHq/GfPA92PiejAdIex4x33oAl9c+xnwQMAFrz2Rb8bUdgg7D/bXdYz7ur53n7k2v6WA6WHsm24qv7RB0FO6v7R7zcX/tnA42OhYFR0AErFXAGI8AEJB1jvtry8V8az2Fg/PvLdNr63ehmK9tuHZ8bQERAO6vrR/zcX/tvACMN6g91zZcOx770QomxeP+2voxX7qftE7A0/YejCeo9lx7omvHYz+SA7TruL+2fszX3h/jCwfn31um15bvQVMuFqxtGN/DrXOsBwWxznF/bb+Yb62nAIwXrDfAgq0N437WubU/zMd6rP3i/to58gx4QbshxtduGPezzrFe6xzrIWBBARSs9twKFPOxPqG9YhnzcX/tHA3bvIOtnwkwXyuYFI+ACoLCBcyX7ietuwMsNTTudQQk8CyWMd/aDwFP25+Dkbj1BOGJLI6scAHzrfX0nY/1Cu0Vy5hvrRd5Nv8ebG2473wEVBAULmC+tV4C5iO6rb9Gaj3RfeejAwXDFsuYb62XDqaD6WCNi9CBhUWFC5ivufdJsXQwHUwHn+SM066hAwXDFsuYf9p9znqdDqaD6zpY+/vc2if6rCf/vHFY77j7HbmDUQDt/LzCjypP248Ub62bgHt+REsApXUCrgzIKqgETLturad3B+PvX61za8N951v7w3xrvb0DthbIfNuHXAJu7BFf+0ATMAHbHhG1TyT30/Ggg+lg3Ymhw9rSiw6mg9s6kXxC6HjQwXSw7sTQYW3pRQfTwW2dSD4hdDzoYDpYd2LosLb0ooPp4LZOJJ8QOh50MB2sOzF0WFt60cF0cFsnkk8IHQ86mA7WnRg6rC296OBpd/Dqu0+Rw68GhYNXXq4f4UXOj//fQ171SGzD8tr60GsDrFs6iOvDcPP+k5mnrzYOKZYklq/1xDSxDWHmwcWr84NLz15v3H7+4csch38NEsvENLH9DwLs1co+Fv2iAAAAAElFTkSuQmCC",ce=""+new URL("Retry-8b2ec896.png",import.meta.url).href,ue=""+new URL("Timeout-dcabf0f7.jpg",import.meta.url).href,pe="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAABYklEQVRoBe2av0oDQRDGZxbRxhfwDRI0NhKtRAhWPkM6Ex9KTOczWElArBRsAuEeIS+QRpvJfJdcqkWRLWYH5or7s7N797v59j4Odph2m4hw//xywsT3JHQqJMddrIajcq2Jaalcs2bx+cTMAi7Grn9xfSI/388kMsJ19RvznA+Pxs3X+yoh867gkV1NNJjBzr3BcKpT5rH6rOcAmR5SO+dzQQdtYE/4YB2w5hGVPdXmNnnSfCvYUz7kpzVewFor9woc/DeDb/OXX4fcjO728b/67jsWnLhXgHtnw/anqCAJpkPdKxAvYDp/9OHhQtYKhAtZKxAuZK1AuJC1AuFC1gqEC1krEC5krUC4kLUC4ULWCoQLWSsQLmStQLhQKFCYAaxSrgvvYTYc7AnL92YEpQ9WdqxSzkrvYzUe7Lwt8rh6dVMn0WVL6yWaxcdtQtUHCidIG7pY9cddsUfL3sF6LbfZAN5wf/+tIkpkAAAAAElFTkSuQmCC",de="data:image/png;base64,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",yt="data:image/png;base64,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",ye="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAAEs0lEQVRoBe1azWtUVxQ/781HkslHzQwpDfkQUUpdaHZtaUtTuhACYtC/wI22FHd+bMSlFNSdIhjcddFNKW0pZlfS0BayEdSFqAjRJMbGfBgTZ+JM8p7nd27ezMvkvsy7yUucAS+898479+P8zj3n3nvembGIy8xMttOhwiVy6RuX3HbwqrVYZE2SRUM2Jc5lMqlxaxX8Hdd109UKWofLsqxZVqLHxszXGngoBMzAbsNtdBrWBI+x29Xu8xtNJLDbGzWohbr3CrxrK8W3A4BtW9SYqqdEQg1fKCzT6+wSOY4bubjIFQD41g+ayLZL3hmrS1KSlZmbX4xciZKUiOYGMw/wfz/M0ldXnsgFGjzURV2sfRceF+3KhwPxCYdDQslhml+ImVq54KKlVwv6v7Pd9GFzXIA/f7Ui/T5qidE/Z3bT1MIyfXn5qfRBhb9/ptGmvv11dOLzFCVi0i3ULe560mVEBr/6lN4igW/+Nr5hRU+u8/TlfdlXqychr9QO8tQUTGcd+ul2TmT98EVK31nDtaamX1aWqukYxGpuaqB69nm4zfk/Xkizi0faqPfjFC29ydPCIoPUlH9H83T61gKlUzbdOt6qaaFnRa6AbhFDtOM4FRfxpzdmxNgj32X0aDXcyHchbJXYbTa1jTIa0502cgXUbLuBrqKZxCJrhffEgP2i2Kac2BYFyoWEfmc0pguyqhRwePusaQu4cW9bDW0z2hYLbDYWcmOmDkTRK6DbRsPGQuJC4SdfWm5bLDQ4sURdv07KBbpSLDT8f55c9oc0hxQmxciFCg7RtUdZ+v1ZnqbfOBIz+WMn0HePdhFCtpMjczSe4w6r9NixdprnffLA4CxzAlwlZlF/d530CXszUuDq4yzdfLqkgh+eKMDwLzuhA+ImAEIc5LCfl3YaNFajtNXZ1N+epFN7w8dBGNNIgd+m8gJgoKeFvs4k0H9daeZEDcrAZ61iBY/GcxfX3T8UPkxAn0rFKBb6ZHhW3OZBbzAI3SIGiDCxUCWwunojCzichPHZXzeefHFtOhbSjrgx00gBHDRhCgK6oLA5TH+TNkYKYAFWWzFSgOSory4VjBRQFgjnRjulptGxJ8FWiA9u7ET4tEy3NssFGrytlO9fLNLMynoXNlOAW1daB942iu/iGKdScIFWuaLNK/FnNk/fTr4kPP3FSIG2es7Gs9P99brgH2MN7eWFBl/lqOv+hFygK8VCawYJeIEFYAm/NYwOsh/ncnR9PldMo3hhgHpCqkWjB7uoPRkX4OMFlRfq5ETP2P4Omswv0557Y3IYKoywiAolpDe/+tNQHi1pm7KpznDcdDHdaBZKnNnVwPGMS78s5mlqhUMGBDUiX7mGCFKkwld+R/PVSwDzrQSf3ZPfMaQKRvCCrBEz+Mm/jaHLumJkgXW9NQwvLwS3OTmByJPjoo409bU0bJgX0gy1htX5RI0F5uFUUmYfVjDaRteMGPCCLDQSuQA81tJRbIVYCHVbKZ7bQAGvRK7AlvJCHirN0z/r/urIXcg/+E7QZWt7J0RGK+O9AtHOp/loHKHwfw9qtAC7zefDUI3i5wOOhmr/zx74ywr+9cE5nZ9rwZ2AEViBGdjfAhPs4mowdpbkAAAAAElFTkSuQmCC",mt=""+new URL("VirtualOutbound-3bb05b70.png",import.meta.url).href,ht={class:"policy-type-tag"},vt=["src"],gt=Y({__name:"PolicyTypeTag",props:{policyType:{type:String,required:!0}},setup(u){const o=u,c=ae(),O={CircuitBreaker:{iconUrl:oe},FaultInjection:{iconUrl:dt},HealthCheck:{iconUrl:le},MeshAccessLog:{iconUrl:pe},MeshCircuitBreaker:{iconUrl:oe},MeshGateway:{iconUrl:null},MeshGatewayRoute:{iconUrl:null},MeshHealthCheck:{iconUrl:le},MeshProxyPatch:{iconUrl:Ae},MeshRateLimit:{iconUrl:re},MeshRetry:{iconUrl:ce},MeshTimeout:{iconUrl:ue},MeshTrace:{iconUrl:ye},MeshTrafficPermission:{iconUrl:de},ProxyTemplate:{iconUrl:Ae},RateLimit:{iconUrl:re},Retry:{iconUrl:ce},Timeout:{iconUrl:ue},TrafficLog:{iconUrl:pe},TrafficPermission:{iconUrl:de},TrafficRoute:{iconUrl:yt},TrafficTrace:{iconUrl:ye},VirtualOutbound:{iconUrl:mt}},D=x(()=>{const T=c.state.policyTypes.map(A=>{const b=O[A.name]??{iconUrl:null};return[A.name,b]});return Object.fromEntries(T)}),m=x(()=>D.value[o.policyType]);return(T,A)=>(e(),s("span",ht,[m.value.iconUrl!==null?(e(),s("img",{key:0,class:"policy-type-tag-icon",src:m.value.iconUrl,alt:""},null,8,vt)):(e(),y(f(me),{key:1,icon:"brain",size:"24"})),a(),Ee(T.$slots,"default",{},()=>[a(d(o.policyType),1)],!0)]))}});const De=K(gt,[["__scopeId","data-v-0052ac03"]]),ft={class:"policy-type-heading"},wt={class:"policy-list"},Tt={key:0},bt=Y({__name:"PolicyTypeEntryList",props:{id:{type:String,required:!1,default:"entry-list"},policyTypeEntries:{type:Object,required:!0}},setup(u){const o=u,c=[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function O({headerKey:D}){return{class:`cell-${D}`}}return(D,m)=>{const T=X("router-link");return e(),y($,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(v,null,Q(o.policyTypeEntries,(A,b)=>(e(),y(Z,{key:b},{"accordion-header":t(()=>[h("h3",ft,[i(De,{"policy-type":A.type},{default:t(()=>[a(d(A.type)+" ("+d(A.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[h("div",wt,[i(f(he),{class:"policy-type-table",fetcher:()=>({data:A.connections,total:A.connections.length}),headers:c,"cell-attrs":O,"disable-pagination":"","is-clickable":""},{sourceTags:t(({rowValue:n})=>[n.length>0?(e(),y(j,{key:0,class:"tag-list",tags:n},null,8,["tags"])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),destinationTags:t(({rowValue:n})=>[n.length>0?(e(),y(j,{key:0,class:"tag-list",tags:n},null,8,["tags"])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),name:t(({rowValue:n})=>[n!==null?(e(),s(v,{key:0},[a(d(n),1)],64)):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),origins:t(({rowValue:n})=>[n.length>0?(e(),s("ul",Tt,[(e(!0),s(v,null,Q(n,(w,U)=>(e(),s("li",{key:`${b}-${U}`},[i(T,{to:w.route},{default:t(()=>[a(d(w.name),1)]),_:2},1032,["to"])]))),128))])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),config:t(({rowValue:n,rowKey:w})=>[n!==null?(e(),y(be,{key:0,id:`${o.id}-${b}-${w}-code-block`,code:n,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const Dt=K(bt,[["__scopeId","data-v-71c85650"]]),Bt={class:"policy-type-heading"},Ct={class:"policy-list"},kt={key:1,class:"tag-list-wrapper"},Ot={key:0},Et={key:1},Pt={key:0},Gt={key:0},Qt=Y({__name:"RuleEntryList",props:{id:{type:String,required:!1,default:"entry-list"},ruleEntries:{type:Object,required:!0}},setup(u){const o=u,c=[{label:"Type",key:"type"},{label:"Addresses",key:"addresses"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function O({headerKey:D}){return{class:`cell-${D}`}}return(D,m)=>{const T=X("router-link");return e(),y($,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(v,null,Q(o.ruleEntries,(A,b)=>(e(),y(Z,{key:b},{"accordion-header":t(()=>[h("h3",Bt,[i(De,{"policy-type":A.type},{default:t(()=>[a(d(A.type)+" ("+d(A.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[h("div",Ct,[i(f(he),{class:"policy-type-table",fetcher:()=>({data:A.connections,total:A.connections.length}),headers:c,"cell-attrs":O,"disable-pagination":"","is-clickable":""},{type:t(({rowValue:n})=>[n.sourceTags.length===0&&n.destinationTags.length===0?(e(),s(v,{key:0},[a(`
                —
              `)],64)):(e(),s("div",kt,[n.sourceTags.length>0?(e(),s("div",Ot,[a(`
                  From

                  `),i(j,{class:"tag-list",tags:n.sourceTags},null,8,["tags"])])):_("",!0),a(),n.destinationTags.length>0?(e(),s("div",Et,[a(`
                  To

                  `),i(j,{class:"tag-list",tags:n.destinationTags},null,8,["tags"])])):_("",!0)]))]),addresses:t(({rowValue:n})=>[n.length>0?(e(),s("ul",Pt,[(e(!0),s(v,null,Q(n,(w,U)=>(e(),s("li",{key:`${b}-${U}`},d(w),1))),128))])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),origins:t(({rowValue:n})=>[n.length>0?(e(),s("ul",Gt,[(e(!0),s(v,null,Q(n,(w,U)=>(e(),s("li",{key:`${b}-${U}`},[i(T,{to:w.route},{default:t(()=>[a(d(w.name),1)]),_:2},1032,["to"])]))),128))])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),config:t(({rowValue:n,rowKey:w})=>[n!==null?(e(),y(be,{key:0,id:`${o.id}-${b}-${w}-code-block`,code:n,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const Ut=K(Qt,[["__scopeId","data-v-74be3da4"]]),It=h("h2",{class:"visually-hidden"},`
    Policies
  `,-1),zt={key:0,class:"mt-2"},Lt=h("h2",{class:"mb-2"},`
      Rules
    `,-1),Mt=Y({__name:"SidecarDataplanePolicyList",props:{dppName:{type:String,required:!0},policyTypeEntries:{type:Object,required:!0},ruleEntries:{type:Array,required:!0}},setup(u){const o=u;return(c,O)=>(e(),s(v,null,[It,a(),i(Dt,{id:"policies","policy-type-entries":o.policyTypeEntries},null,8,["policy-type-entries"]),a(),u.ruleEntries.length>0?(e(),s("div",zt,[Lt,a(),i(Ut,{id:"rules","rule-entries":o.ruleEntries},null,8,["rule-entries"])])):_("",!0)],64))}}),Rt={key:2,class:"policies-list"},xt={key:3,class:"policies-list"},_t=Y({__name:"DataplanePolicies",props:{dataplaneOverview:{type:Object,required:!0}},setup(u){const o=u,c=ve(),O=ae(),D=S(null),m=S([]),T=S([]),A=S([]),b=S([]),n=S(!0),w=S(null);Pe(()=>o.dataplaneOverview.name,function(){U()}),U();async function U(){var r,p;w.value=null,n.value=!0,m.value=[],T.value=[],A.value=[],b.value=[];try{if(((p=(r=o.dataplaneOverview.dataplane.networking.gateway)==null?void 0:r.type)==null?void 0:p.toUpperCase())==="BUILTIN")D.value=await c.getMeshGatewayDataplane({mesh:o.dataplaneOverview.mesh,name:o.dataplaneOverview.name}),A.value=J(D.value),b.value=W(D.value.policies);else{const{items:l}=await c.getSidecarDataplanePolicies({mesh:o.dataplaneOverview.mesh,name:o.dataplaneOverview.name});m.value=I(l??[]);const{items:B}=await c.getDataplaneRules({mesh:o.dataplaneOverview.mesh,name:o.dataplaneOverview.name});T.value=L(B??[])}}catch(g){g instanceof Error?w.value=g:console.error(g)}finally{n.value=!1}}function J(r){const p=[],g=r.listeners??[];for(const l of g)for(const B of l.hosts)for(const k of B.routes){const P=[];for(const G of k.destinations){const C=W(G.policies),R={routeName:k.route,route:{name:"policy-detail-view",params:{mesh:r.gateway.mesh,policyPath:"meshgatewayroutes",policy:k.route}},service:G.tags["kuma.io/service"],policies:C};P.push(R)}p.push({protocol:l.protocol,port:l.port,hostName:B.hostName,routeEntries:P})}return p}function W(r){if(r===void 0)return[];const p=[];for(const g of Object.values(r)){const l=O.state.policyTypesByName[g.type];p.push({type:g.type,name:g.name,route:{name:"policy-detail-view",params:{mesh:g.mesh,policyPath:l.path,policy:g.name}}})}return p}function I(r){const p=new Map;for(const l of r){const{type:B,service:k}=l,P=typeof k=="string"&&k!==""?[{label:"kuma.io/service",value:k}]:[],G=B==="inbound"||B==="outbound"?l.name:null;for(const[C,R]of Object.entries(l.matchedPolicies)){p.has(C)||p.set(C,{type:C,connections:[]});const H=p.get(C),q=O.state.policyTypesByName[C];for(const se of R){const M=z(se,q,l,P,G);H.connections.push(...M)}}}const g=Array.from(p.values());return g.sort((l,B)=>l.type.localeCompare(B.type)),g}function z(r,p,g,l,B){const k=r.conf&&Object.keys(r.conf).length>0?ne(r.conf):null,G=[{name:r.name,route:{name:"policy-detail-view",params:{mesh:r.mesh,policyPath:p.path,policy:r.name}}}],C=[];if(g.type==="inbound"&&Array.isArray(r.sources))for(const{match:R}of r.sources){const q={sourceTags:[{label:"kuma.io/service",value:R["kuma.io/service"]}],destinationTags:l,name:B,config:k,origins:G};C.push(q)}else{const H={sourceTags:[],destinationTags:l,name:B,config:k,origins:G};C.push(H)}return C}function L(r){const p=new Map;for(const l of r){p.has(l.policyType)||p.set(l.policyType,{type:l.policyType,connections:[]});const B=p.get(l.policyType),k=O.state.policyTypesByName[l.policyType],P=E(l,k);B.connections.push(...P)}const g=Array.from(p.values());return g.sort((l,B)=>l.type.localeCompare(B.type)),g}function E(r,p){const{type:g,service:l,subset:B,conf:k}=r,P=B?Object.entries(B):[];let G,C;g==="ClientSubset"?P.length>0?G=P.map(([M,ee])=>({label:M,value:ee})):G=[{label:"kuma.io/service",value:"*"}]:G=[],g==="DestinationSubset"?P.length>0?C=P.map(([M,ee])=>({label:M,value:ee})):typeof l=="string"&&l!==""?C=[{label:"kuma.io/service",value:l}]:C=[{label:"kuma.io/service",value:"*"}]:g==="ClientSubset"&&typeof l=="string"&&l!==""?C=[{label:"kuma.io/service",value:l}]:C=[];const R=r.addresses??[],H=k&&Object.keys(k).length>0?ne(k):null,q=[];for(const M of r.origins)q.push({name:M.name,route:{name:"policy-detail-view",params:{mesh:M.mesh,policyPath:p.path,policy:M.name}}});return[{type:{sourceTags:G,destinationTags:C},addresses:R,config:H,origins:q}]}return(r,p)=>n.value?(e(),y(ge,{key:0})):w.value!==null?(e(),y(fe,{key:1,error:w.value},null,8,["error"])):m.value.length>0?(e(),s("div",Rt,[i(Mt,{"dpp-name":o.dataplaneOverview.name,"policy-type-entries":m.value,"rule-entries":T.value},null,8,["dpp-name","policy-type-entries","rule-entries"])])):A.value.length>0&&D.value!==null?(e(),s("div",xt,[i(pt,{"mesh-gateway-dataplane":D.value,"mesh-gateway-listener-entries":A.value,"mesh-gateway-route-policies":b.value},null,8,["mesh-gateway-dataplane","mesh-gateway-listener-entries","mesh-gateway-route-policies"])])):(e(),y(we,{key:4}))}});const Yt=K(_t,[["__scopeId","data-v-2a40d801"]]),St={class:"stack"},Nt={class:"variable-columns"},Ht={class:"status-with-reason"},qt=["href"],jt=Y({__name:"DataPlaneDetails",props:{dataplaneOverview:{type:Object,required:!0}},setup(u){const o=u,{t:c,formatIsoDate:O}=Te(),D=ve(),m=ae(),T=[{hash:"#overview",title:c("data-planes.routes.item.tabs.overview")},{hash:"#insights",title:c("data-planes.routes.item.tabs.insights")},{hash:"#dpp-policies",title:c("data-planes.routes.item.tabs.policies")},{hash:"#xds-configuration",title:c("data-planes.routes.item.tabs.xds_configuration")},{hash:"#envoy-stats",title:c("data-planes.routes.item.tabs.stats")},{hash:"#envoy-clusters",title:c("data-planes.routes.item.tabs.clusters")},{hash:"#mtls",title:c("data-planes.routes.item.tabs.mtls")}],A=x(()=>Ke(o.dataplaneOverview.dataplane,o.dataplaneOverview.dataplaneInsight)),b=x(()=>ie(o.dataplaneOverview.dataplane)),n=x(()=>Fe(o.dataplaneOverview.dataplaneInsight)),w=x(()=>Je(o.dataplaneOverview,O)),U=x(()=>{var z;const I=Array.from(((z=o.dataplaneOverview.dataplaneInsight)==null?void 0:z.subscriptions)??[]);return I.reverse(),I}),J=x(()=>{var r;const I=((r=o.dataplaneOverview.dataplaneInsight)==null?void 0:r.subscriptions)??[];if(I.length===0)return[];const z=I[I.length-1];if(!("version"in z)||!z.version)return[];const L=[],E=z.version;if(E.kumaDp&&E.envoy){const p=We(E);p.kind!==Ve&&p.kind!==Xe&&L.push(p)}return m.getters["config/getMulticlusterStatus"]&&ie(o.dataplaneOverview.dataplane).find(l=>l.label===Ge)&&typeof E.kumaDp.kumaCpCompatible=="boolean"&&!E.kumaDp.kumaCpCompatible&&L.push({kind:Ze,payload:{kumaDp:E.kumaDp.version}}),L});async function W(I){const{mesh:z,name:L}=o.dataplaneOverview;return await D.getDataplaneFromMesh({mesh:z,name:L},I)}return(I,z)=>{const L=X("RouterLink");return e(),y(He,{tabs:T},{overview:t(()=>[h("div",St,[J.value.length>0?(e(),y(je,{key:0,warnings:J.value,"data-testid":"data-plane-warnings"},null,8,["warnings"])):_("",!0),a(),i(f(N),null,{body:t(()=>[h("div",Nt,[i(V,null,{title:t(()=>[a(d(f(c)("http.api.property.status")),1)]),body:t(()=>[h("div",Ht,[i(Le,{status:A.value.status},null,8,["status"]),a(),A.value.reason.length>0?(e(),y(f(Qe),{key:0,label:A.value.reason.join(", "),class:"reason-tooltip"},{default:t(()=>[i(f(me),{icon:"info",size:"20","hide-title":""})]),_:1},8,["label"])):_("",!0)])]),_:1}),a(),i(V,null,{title:t(()=>[a(d(f(c)("http.api.property.name")),1)]),body:t(()=>[i(qe,{text:o.dataplaneOverview.name},{default:t(()=>[i(L,{to:{name:"data-plane-detail-view",params:{mesh:o.dataplaneOverview.mesh,dataPlane:o.dataplaneOverview.name}}},{default:t(()=>[a(d(o.dataplaneOverview.name),1)]),_:1},8,["to"])]),_:1},8,["text"])]),_:1}),a(),i(V,null,{title:t(()=>[a(d(f(c)("http.api.property.tags")),1)]),body:t(()=>[b.value.length>0?(e(),y(j,{key:0,tags:b.value},null,8,["tags"])):(e(),s(v,{key:1},[a(d(f(c)("common.detail.none")),1)],64))]),_:1}),a(),i(V,null,{title:t(()=>[a(d(f(c)("http.api.property.dependencies")),1)]),body:t(()=>[n.value!==null?(e(),y(j,{key:0,tags:n.value},null,8,["tags"])):(e(),s(v,{key:1},[a(d(f(c)("common.detail.none")),1)],64))]),_:1})])]),_:1}),a(),i(_e,{id:"code-block-data-plane",resource:o.dataplaneOverview,"resource-fetcher":W,"is-searchable":""},null,8,["resource"])])]),insights:t(()=>[i(f(N),null,{body:t(()=>[i(Ne,{"is-empty":U.value.length===0},{default:t(()=>[i($,{"initially-open":0},{default:t(()=>[(e(!0),s(v,null,Q(U.value,(E,r)=>(e(),y(Z,{key:r},{"accordion-header":t(()=>[i(Ie,{subscription:E},null,8,["subscription"])]),"accordion-content":t(()=>[i(ze,{subscription:E,"is-discovery-subscription":""},null,8,["subscription"])]),_:2},1024))),128))]),_:1})]),_:1},8,["is-empty"])]),_:1})]),"dpp-policies":t(()=>[i(f(N),null,{body:t(()=>[i(Yt,{"dataplane-overview":u.dataplaneOverview},null,8,["dataplane-overview"])]),_:1})]),"xds-configuration":t(()=>[i(f(N),null,{body:t(()=>[i(te,{"data-path":"xds",mesh:u.dataplaneOverview.mesh,"dpp-name":u.dataplaneOverview.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),_:1})]),"envoy-stats":t(()=>[i(f(N),null,{body:t(()=>[i(te,{"data-path":"stats",mesh:u.dataplaneOverview.mesh,"dpp-name":u.dataplaneOverview.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),_:1})]),"envoy-clusters":t(()=>[i(f(N),null,{body:t(()=>[i(te,{"data-path":"clusters",mesh:u.dataplaneOverview.mesh,"dpp-name":u.dataplaneOverview.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),_:1})]),mtls:t(()=>[i(f(N),null,{body:t(()=>[w.value===null?(e(),y(f(Ue),{key:0,appearance:"danger"},{alertMessage:t(()=>[a(`
              This data plane proxy does not yet have mTLS configured —
              `),h("a",{href:f(c)("data-planes.href.docs.mutual-tls"),class:"external-link",target:"_blank"},`
                Learn About Certificates in `+d(f(c)("common.product.name")),9,qt)]),_:1})):(e(),y(Ye,{key:1},{default:t(()=>[(e(!0),s(v,null,Q(w.value,(E,r)=>(e(),y(Se,{key:r,term:f(c)(`http.api.property.${r}`)},{default:t(()=>[a(d(E),1)]),_:2},1032,["term"]))),128))]),_:1}))]),_:1})]),_:1})}}});const Kt=K(jt,[["__scopeId","data-v-56d4ace4"]]),la=Y({__name:"DataPlaneDetailView",props:{isGatewayView:{type:Boolean,required:!1,default:!1}},setup(u){const o=u,{t:c}=Te();return(O,D)=>(e(),y(xe,{name:"data-plane-detail-view","data-testid":"data-plane-detail-view"},{default:t(({route:m})=>[i(Me,{breadcrumbs:[{to:{name:`${o.isGatewayView?"gateways":"data-planes"}-list-view`,params:{mesh:m.params.mesh}},text:f(c)(`${o.isGatewayView?"gateways":"data-planes"}.routes.item.breadcrumbs`)}]},{title:t(()=>[h("h2",null,[i($e,{title:f(c)(`${o.isGatewayView?"gateways":"data-planes"}.routes.item.title`,{name:m.params.dataPlane}),render:!0},null,8,["title"])])]),default:t(()=>[a(),i(Re,{src:`/meshes/${m.params.mesh}/dataplane-overviews/${m.params.dataPlane}`},{default:t(({data:T,isLoading:A,error:b})=>[A?(e(),y(ge,{key:0})):b?(e(),y(fe,{key:1,error:b},null,8,["error"])):T===void 0?(e(),y(we,{key:2})):(e(),y(Kt,{key:3,"dataplane-overview":T,"data-testid":"detail-view-details"},null,8,["dataplane-overview"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{la as default};
